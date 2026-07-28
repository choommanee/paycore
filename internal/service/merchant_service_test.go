package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/crypto"
	"github.com/yourco/payment-gateway/internal/repository"
)

// fakeMerchantRepo implements the merchant slice of repository.Querier.
type fakeMerchantRepo struct {
	repository.Querier

	mu        sync.Mutex
	byID      map[uuid.UUID]repository.Merchant
	byHash    map[string]repository.Merchant
	createErr error

	// seriesRows backs StatsSeriesByDay: every call returns the subset of these
	// rows whose day falls in [arg.CreatedAt, arg.CreatedAt_2), mimicking the SQL
	// WHERE clause so a test can seed rows for both the current and previous
	// windows in one slice.
	seriesRows []repository.StatsSeriesByDayRow
	seriesErr  error
}

func newFakeMerchantRepo() *fakeMerchantRepo {
	return &fakeMerchantRepo{
		byID:   map[uuid.UUID]repository.Merchant{},
		byHash: map[string]repository.Merchant{},
	}
}

func (f *fakeMerchantRepo) CreateMerchant(_ context.Context, arg repository.CreateMerchantParams) (repository.Merchant, error) {
	if f.createErr != nil {
		return repository.Merchant{}, f.createErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	m := repository.Merchant{
		ID:                 arg.ID,
		Name:               arg.Name,
		Status:             "active",
		ApiKeyHash:         arg.ApiKeyHash,
		Mcc:                arg.Mcc,
		SettlementCurrency: arg.SettlementCurrency,
	}
	f.byID[key(arg.ID)] = m
	f.byHash[arg.ApiKeyHash] = m
	return m, nil
}

func (f *fakeMerchantRepo) GetMerchant(_ context.Context, id pgtype.UUID) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byID[key(id)]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return m, nil
}

func (f *fakeMerchantRepo) GetMerchantByAPIKeyHash(_ context.Context, hash string) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byHash[hash]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return m, nil
}

func (f *fakeMerchantRepo) StatsSeriesByDay(_ context.Context, arg repository.StatsSeriesByDayParams) ([]repository.StatsSeriesByDayRow, error) {
	if f.seriesErr != nil {
		return nil, f.seriesErr
	}
	from, to := arg.CreatedAt.Time, arg.CreatedAt_2.Time
	var out []repository.StatsSeriesByDayRow
	for _, r := range f.seriesRows {
		d := r.Day.Time
		if !d.Before(from) && d.Before(to) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeMerchantRepo) SetMerchantWebhook(_ context.Context, arg repository.SetMerchantWebhookParams) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byID[key(arg.ID)]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	m.WebhookUrl = arg.WebhookUrl
	m.WebhookSecretHash = arg.WebhookSecretHash
	m.WebhookSecretEnc = arg.WebhookSecretEnc
	f.byID[key(arg.ID)] = m
	return m, nil
}

// ---- SetWebhook: per-merchant secret stored retrievably (Fix 2) ------------

// TestSetWebhookStoresEncryptedSecret proves SetWebhook persists the signing
// secret as an envelope-encrypted blob that decrypts back to the whsec_ returned
// to the merchant (NOT a one-way hash), so the outbound worker can sign this
// merchant's deliveries with the merchant's own key.
func TestSetWebhookStoresEncryptedSecret(t *testing.T) {
	repo := newFakeMerchantRepo()
	kms, err := crypto.NewDevLocalKMS("dev-key")
	if err != nil {
		t.Fatalf("kms: %v", err)
	}
	box := crypto.NewSecretBox(kms)
	svc := NewMerchantService(repo, zerolog.Nop()).(*merchantService).WithWebhookSecrets(box)

	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatalf("Onboard: %v", err)
	}
	cfg, err := svc.SetWebhook(context.Background(), cred.Merchant.ID, "https://merchant.example.com/hook")
	if err != nil {
		t.Fatalf("SetWebhook: %v", err)
	}
	if !strings.HasPrefix(cfg.SigningSecret, "whsec_") {
		t.Fatalf("signing secret has wrong prefix: %q", cfg.SigningSecret)
	}

	stored := repo.byID[cred.Merchant.ID]
	if len(stored.WebhookSecretEnc) == 0 {
		t.Fatal("webhook_secret_enc was not persisted")
	}
	if strings.Contains(string(stored.WebhookSecretEnc), cfg.SigningSecret) {
		t.Fatal("stored blob leaks the raw secret (must be encrypted)")
	}
	// The stored blob must decrypt back to the exact secret handed to the merchant.
	plain, err := box.Open(context.Background(), stored.WebhookSecretEnc)
	if err != nil {
		t.Fatalf("Open stored secret: %v", err)
	}
	if string(plain) != cfg.SigningSecret {
		t.Fatalf("decrypted secret %q != returned secret %q", plain, cfg.SigningSecret)
	}
}

// TestSetWebhookNoSealerStoresHashOnly asserts backward compatibility: without a
// sealer, no encrypted secret is written (column stays NULL) and the worker will
// fall back to the global key.
func TestSetWebhookNoSealerStoresHashOnly(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatalf("Onboard: %v", err)
	}
	if _, err := svc.SetWebhook(context.Background(), cred.Merchant.ID, "https://m.example.com/hook"); err != nil {
		t.Fatalf("SetWebhook: %v", err)
	}
	stored := repo.byID[cred.Merchant.ID]
	if len(stored.WebhookSecretEnc) != 0 {
		t.Fatal("expected no encrypted secret when no sealer is configured")
	}
	if stored.WebhookSecretHash == nil || *stored.WebhookSecretHash == "" {
		t.Fatal("expected the secret hash to still be persisted")
	}
}

// ---- Onboard: key generation + hash-only persistence ----------------------

func TestOnboardIssuesKeyAndPersistsHashOnly(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())

	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{
		Name: "Acme", SettlementCurrency: "THB",
	})
	if err != nil {
		t.Fatalf("Onboard: %v", err)
	}
	if !strings.HasPrefix(cred.APIKey, "sk_live_") {
		t.Fatalf("api key has wrong prefix: %q", cred.APIKey)
	}
	// The stored hash must be the SHA-256 of the issued key, and the raw key must
	// NOT be persisted anywhere.
	stored, ok := repo.byHash[middleware.HashAPIKey(cred.APIKey)]
	if !ok {
		t.Fatal("stored merchant is not keyed by the hash of the issued API key")
	}
	if stored.ApiKeyHash == cred.APIKey {
		t.Fatal("raw API key was persisted instead of its hash")
	}
	if stored.ApiKeyHash == "" || len(stored.ApiKeyHash) != 64 {
		t.Fatalf("stored hash looks wrong: %q", stored.ApiKeyHash)
	}
}

func TestOnboardNoRepoErrors(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "X", SettlementCurrency: "THB"}); err == nil {
		t.Fatal("expected error when repo is not configured")
	}
}

func TestOnboardPropagatesRepoError(t *testing.T) {
	repo := newFakeMerchantRepo()
	repo.createErr = errors.New("db down")
	svc := NewMerchantService(repo, zerolog.Nop())
	if _, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "X", SettlementCurrency: "THB"}); err == nil {
		t.Fatal("expected the repo error to propagate")
	}
}

// ---- Get ------------------------------------------------------------------

func TestMerchantGetRoundTrip(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), cred.Merchant.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Acme" || got.SettlementCurrency != "THB" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestMerchantGetNotFound(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	if _, err := svc.Get(context.Background(), uuid.New()); !errors.Is(err, domain.ErrMerchantNotFound) {
		t.Fatalf("err=%v want ErrMerchantNotFound", err)
	}
}

func TestMerchantGetNoRepo(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.Get(context.Background(), uuid.New()); !errors.Is(err, domain.ErrMerchantNotFound) {
		t.Fatalf("err=%v want ErrMerchantNotFound", err)
	}
}

// ---- ResolveByAPIKeyHash (auth resolution) --------------------------------

func TestResolveByAPIKeyHashSuccessAndMiss(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatal(err)
	}

	// Correct hash resolves the merchant.
	m, err := svc.ResolveByAPIKeyHash(context.Background(), middleware.HashAPIKey(cred.APIKey))
	if err != nil {
		t.Fatalf("resolve valid: %v", err)
	}
	if m.ID != cred.Merchant.ID {
		t.Fatalf("resolved wrong merchant: %s != %s", m.ID, cred.Merchant.ID)
	}

	// An unknown hash is unauthorized (never leaks existence).
	if _, err := svc.ResolveByAPIKeyHash(context.Background(), middleware.HashAPIKey("sk_live_bogus")); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("err=%v want ErrUnauthorized", err)
	}
}

func TestResolveByAPIKeyHashNoRepo(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.ResolveByAPIKeyHash(context.Background(), "x"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("err=%v want ErrUnauthorized", err)
	}
}

// Two onboardings must never mint the same API key (crypto-random).
func TestGenerateAPIKeyUnique(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	c1, _ := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "A", SettlementCurrency: "THB"})
	c2, _ := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "B", SettlementCurrency: "THB"})
	if c1.APIKey == c2.APIKey {
		t.Fatal("two onboardings minted the same API key")
	}
}

// ---- StatsSeries ------------------------------------------------------------

// utcDay truncates t to a UTC calendar day (00:00:00), matching how the
// service and the (real) SQL query bucket rows.
func utcDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// TestStatsSeriesZeroFillsAndComputesTotals seeds two days of rows inside a
// 3-day window (with a gap day that must be zero-filled) plus a row that only
// falls inside the previous equal-length window, and asserts: the series is
// ascending with one point per calendar day, the gap day is zero-filled,
// totals sum the window, and trend compares against the previous window.
func TestStatsSeriesZeroFillsAndComputesTotals(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	mid := uuid.New()

	today := utcDay(time.Now())
	day := func(offset int) time.Time { return today.AddDate(0, 0, offset) }

	repo.seriesRows = []repository.StatsSeriesByDayRow{
		// Current window (last 3 days: today-2, today-1, today). today-1 is left
		// out on purpose so it must come back zero-filled.
		{Day: pgtype.Date{Time: day(-2), Valid: true}, VolumeMinor: 10000, Count: 2},
		{Day: pgtype.Date{Time: day(0), Valid: true}, VolumeMinor: 5000, Count: 1},
		// Falls only inside the previous 3-day window ([today-5, today-2)).
		{Day: pgtype.Date{Time: day(-4), Valid: true}, VolumeMinor: 6000, Count: 3},
	}

	got, err := svc.StatsSeries(context.Background(), mid, 3)
	if err != nil {
		t.Fatalf("StatsSeries: %v", err)
	}
	if got.Days != 3 {
		t.Fatalf("Days=%d want 3", got.Days)
	}
	if len(got.Series) != 3 {
		t.Fatalf("len(Series)=%d want 3: %+v", len(got.Series), got.Series)
	}

	wantDates := []string{day(-2).Format("2006-01-02"), day(-1).Format("2006-01-02"), day(0).Format("2006-01-02")}
	for i, p := range got.Series {
		if p.Date != wantDates[i] {
			t.Fatalf("Series[%d].Date=%q want %q (ascending order)", i, p.Date, wantDates[i])
		}
	}
	if got.Series[0].VolumeMinor != 10000 || got.Series[0].Count != 2 {
		t.Fatalf("Series[0]=%+v want {volume 10000 count 2}", got.Series[0])
	}
	if got.Series[1].VolumeMinor != 0 || got.Series[1].Count != 0 {
		t.Fatalf("Series[1] (gap day) = %+v want zero-filled", got.Series[1])
	}
	if got.Series[2].VolumeMinor != 5000 || got.Series[2].Count != 1 {
		t.Fatalf("Series[2]=%+v want {volume 5000 count 1}", got.Series[2])
	}

	if got.Totals.VolumeMinor != 15000 || got.Totals.Count != 3 {
		t.Fatalf("Totals=%+v want {volume 15000 count 3}", got.Totals)
	}

	// prev window totals: volume 6000, count 3.
	wantVolumePct := float64(15000-6000) / float64(6000)
	if got.Trend.VolumePct != wantVolumePct {
		t.Fatalf("Trend.VolumePct=%v want %v", got.Trend.VolumePct, wantVolumePct)
	}
	if got.Trend.CountPct != 0 { // 3 vs 3 -> unchanged
		t.Fatalf("Trend.CountPct=%v want 0 (3 vs 3, unchanged)", got.Trend.CountPct)
	}
}

// TestStatsSeriesPrevWindowZeroTrendIsZero asserts the divide-by-zero guard:
// when the previous window has no activity, trend is 0 (not +Inf/NaN) even
// though the current window has volume.
func TestStatsSeriesPrevWindowZeroTrendIsZero(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	mid := uuid.New()

	repo.seriesRows = []repository.StatsSeriesByDayRow{
		{Day: pgtype.Date{Time: utcDay(time.Now()), Valid: true}, VolumeMinor: 1000, Count: 1},
	}

	got, err := svc.StatsSeries(context.Background(), mid, 5)
	if err != nil {
		t.Fatalf("StatsSeries: %v", err)
	}
	if got.Totals.VolumeMinor != 1000 || got.Totals.Count != 1 {
		t.Fatalf("Totals=%+v want {volume 1000 count 1}", got.Totals)
	}
	if got.Trend.VolumePct != 0 || got.Trend.CountPct != 0 {
		t.Fatalf("Trend=%+v want zero (no prior-window activity to compare against)", got.Trend)
	}
}

// TestStatsSeriesDaysClamp asserts days is clamped to [1, 90], defaulting to
// 30 for a non-positive input.
func TestStatsSeriesDaysClamp(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	mid := uuid.New()

	if got, err := svc.StatsSeries(context.Background(), mid, 0); err != nil || got.Days != 30 || len(got.Series) != 30 {
		t.Fatalf("days=0 -> Days=%d len(Series)=%d err=%v want 30/30", got.Days, len(got.Series), err)
	}
	if got, err := svc.StatsSeries(context.Background(), mid, 500); err != nil || got.Days != 90 || len(got.Series) != 90 {
		t.Fatalf("days=500 -> Days=%d len(Series)=%d err=%v want 90/90", got.Days, len(got.Series), err)
	}
	if got, err := svc.StatsSeries(context.Background(), mid, -1); err != nil || got.Days != 30 {
		t.Fatalf("days=-1 -> Days=%d err=%v want 30", got.Days, err)
	}
}

// TestStatsSeriesNoRepo mirrors TestOnboardNoRepoErrors / TestMerchantGetNoRepo:
// an unconfigured repo must not panic, returning a harmless empty result.
func TestStatsSeriesNoRepo(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	got, err := svc.StatsSeries(context.Background(), uuid.New(), 30)
	if err != nil {
		t.Fatalf("StatsSeries: %v", err)
	}
	if got.Days != 30 || len(got.Series) != 0 {
		t.Fatalf("got=%+v want Days=30, empty Series", got)
	}
}

// TestStatsSeriesPropagatesRepoError asserts a repo failure on the current
// window surfaces to the caller rather than being swallowed.
func TestStatsSeriesPropagatesRepoError(t *testing.T) {
	repo := newFakeMerchantRepo()
	repo.seriesErr = errors.New("db down")
	svc := NewMerchantService(repo, zerolog.Nop())
	if _, err := svc.StatsSeries(context.Background(), uuid.New(), 30); err == nil {
		t.Fatal("expected the repo error to propagate")
	}
}
