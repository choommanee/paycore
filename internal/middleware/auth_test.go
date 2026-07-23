package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
)

// stubResolver is a MerchantResolver whose behaviour is driven per-test: it
// returns `merchant` when the presented hash matches `wantHash`, otherwise the
// configured `err` (or ErrUnauthorized).
type stubResolver struct {
	wantHash string
	merchant *domain.Merchant
	err      error
}

func (s stubResolver) ResolveByAPIKeyHash(_ context.Context, hash string) (*domain.Merchant, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.merchant != nil && hash == s.wantHash {
		return s.merchant, nil
	}
	return nil, domain.ErrUnauthorized
}

// mount wires APIKeyAuth in front of a probe handler that echoes the resolved
// merchant id, so a test can assert both the status and that the identity was
// propagated into locals.
func mountAuth(resolver MerchantResolver) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Get("/probe", APIKeyAuth(resolver), func(c *fiber.Ctx) error {
		id, ok := MerchantIDFromCtx(c)
		if !ok {
			return domain.Error(c, fiber.StatusInternalServerError, "NO_ID", "no merchant id")
		}
		return domain.Success(c, fiber.Map{"merchant_id": id.String()})
	})
	return app
}

func doAuth(t *testing.T, app *fiber.App, header, value string) (int, domain.APIResponse) {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodGet, "/probe", nil)
	if header != "" {
		req.Header.Set(header, value)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	_ = json.Unmarshal(body, &env)
	return resp.StatusCode, env
}

func TestAPIKeyAuthMissingKeyRejected(t *testing.T) {
	app := mountAuth(stubResolver{})
	status, env := doAuth(t, app, "", "")
	if status != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", status)
	}
	if env.Code != "UNAUTHORIZED" {
		t.Fatalf("code=%q want UNAUTHORIZED", env.Code)
	}
}

func TestAPIKeyAuthWrongKeyRejected(t *testing.T) {
	m := &domain.Merchant{ID: uuid.New(), Status: "active"}
	app := mountAuth(stubResolver{wantHash: HashAPIKey("sk_live_correct"), merchant: m})

	status, env := doAuth(t, app, "Authorization", "Bearer sk_live_wrong")
	if status != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", status)
	}
	if env.Code != "UNAUTHORIZED" {
		t.Fatalf("code=%q want UNAUTHORIZED", env.Code)
	}
}

// A backend error (e.g. DB down) must still surface as 401 and never leak the
// cause in the response body.
func TestAPIKeyAuthBackendErrorRejectedWithoutLeak(t *testing.T) {
	app := mountAuth(stubResolver{err: context.DeadlineExceeded})
	status, env := doAuth(t, app, "X-API-Key", "sk_live_anything")
	if status != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", status)
	}
	if env.Message == context.DeadlineExceeded.Error() {
		t.Fatalf("auth leaked the backend error cause: %q", env.Message)
	}
}

func TestAPIKeyAuthBearerSucceeds(t *testing.T) {
	m := &domain.Merchant{ID: uuid.New(), Status: "active"}
	app := mountAuth(stubResolver{wantHash: HashAPIKey("sk_live_ok"), merchant: m})

	status, env := doAuth(t, app, "Authorization", "Bearer sk_live_ok")
	if status != fiber.StatusOK {
		t.Fatalf("status=%d want 200", status)
	}
	data, _ := env.Data.(map[string]any)
	if data == nil || data["merchant_id"] != m.ID.String() {
		t.Fatalf("resolved merchant id not propagated: %+v", env.Data)
	}
}

// A bare (non-Bearer) Authorization value and the X-API-Key header are both
// accepted extraction sources.
func TestAPIKeyAuthBareAndXAPIKeyAccepted(t *testing.T) {
	m := &domain.Merchant{ID: uuid.New(), Status: "active"}
	app := mountAuth(stubResolver{wantHash: HashAPIKey("sk_live_bare"), merchant: m})

	if status, _ := doAuth(t, app, "Authorization", "sk_live_bare"); status != fiber.StatusOK {
		t.Fatalf("bare Authorization -> %d, want 200", status)
	}
	if status, _ := doAuth(t, app, "X-API-Key", "sk_live_bare"); status != fiber.StatusOK {
		t.Fatalf("X-API-Key -> %d, want 200", status)
	}
}

func TestHashAPIKeyStableAndOpaque(t *testing.T) {
	h1 := HashAPIKey("sk_live_x")
	h2 := HashAPIKey("sk_live_x")
	if h1 != h2 {
		t.Fatal("HashAPIKey is not deterministic")
	}
	if h1 == "sk_live_x" || len(h1) != 64 {
		t.Fatalf("hash looks wrong (len=%d, val=%q)", len(h1), h1)
	}
}
