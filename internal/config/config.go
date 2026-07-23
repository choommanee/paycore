package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Config holds runtime configuration. Secrets (DB DSN, JWT secret, KMS key IDs)
// should come from environment variables / a secrets manager, never from a
// committed file. Card data encryption keys live in an HSM/KMS, not here.
type Config struct {
	Env         string `mapstructure:"ENV"`
	HTTPAddr    string `mapstructure:"HTTP_ADDR"`
	LogLevelStr string `mapstructure:"LOG_LEVEL"`

	// Optional TLS. When both TLSCertFile and TLSKeyFile are set the server
	// listens over TLS; otherwise it serves plain HTTP (terminate TLS at a load
	// balancer / service mesh in that case).
	TLSCertFile string `mapstructure:"TLS_CERT_FILE"`
	TLSKeyFile  string `mapstructure:"TLS_KEY_FILE"`

	// TLSTerminatedUpstream is an explicit operator acknowledgement that TLS is
	// terminated by a trusted upstream (load balancer / service mesh) in front of
	// this process. It is the ONLY way to run plaintext HTTP when ENV=production;
	// otherwise the server refuses to start. This prevents silently serving a
	// card-acquiring API over cleartext in prod.
	TLSTerminatedUpstream bool `mapstructure:"TLS_TERMINATED_UPSTREAM"`

	// CORS allowlist. Comma-separated list of explicit checkout origins allowed to
	// call the API from a browser. Never use '*' on an authenticated money API.
	// Empty (dev default) disables CORS entirely rather than allowing all origins.
	CORSAllowOrigins string `mapstructure:"CORS_ALLOW_ORIGINS"`

	// Per-merchant rate limit ceiling (requests per second) applied to the
	// authenticated payments/QR route groups. Raised well above real traffic so
	// legitimate bursts and load tests run unthrottled; health/metrics/webhook are
	// never limited.
	RateLimitPerSec int `mapstructure:"RATE_LIMIT_PER_SEC"`

	// SignupRateLimitPerHour caps how many merchant onboarding calls (POST
	// /v1/merchants) a single client IP may make per hour. The endpoint is public
	// and unauthenticated (self-service sandbox signup), so it gets its own
	// dedicated IP-keyed limiter to blunt abuse. Default 5.
	SignupRateLimitPerHour int `mapstructure:"SIGNUP_RATE_LIMIT_PER_HOUR"`

	// MigrateOnBoot, when true, runs all pending up migrations (embedded, in
	// process) against DATABASE_URL before the server starts serving. Off by
	// default for local dev; set MIGRATE_ON_BOOT=true on a platform like Railway
	// whose distroless runtime image has no shell or migrate binary.
	MigrateOnBoot bool `mapstructure:"MIGRATE_ON_BOOT"`

	// Pool sizing (see internal/db.PoolConfig). Tunable per environment.
	DBMaxConns int32 `mapstructure:"DB_MAX_CONNS"`
	DBMinConns int32 `mapstructure:"DB_MIN_CONNS"`

	// MetricsAddr, when set (e.g. ":9090"), binds /metrics on a separate internal
	// listener instead of the public app so business telemetry is not exposed on
	// the money API. Empty keeps /metrics off the public listener entirely unless
	// MetricsPublic is true.
	MetricsAddr   string `mapstructure:"METRICS_ADDR"`
	MetricsPublic bool   `mapstructure:"METRICS_PUBLIC"`

	DatabaseURL string `mapstructure:"DATABASE_URL"`

	JWTSecret       string `mapstructure:"JWT_SECRET"`
	IdempotencyTTLs int    `mapstructure:"IDEMPOTENCY_TTL_SECONDS"`

	// KMS / HSM references used by the tokenization vault.
	KMSKeyID       string `mapstructure:"KMS_KEY_ID"`
	VaultNamespace string `mapstructure:"VAULT_NAMESPACE"`

	// Upstream acquirer / card-network switch.
	AcquirerBaseURL string `mapstructure:"ACQUIRER_BASE_URL"`
	AcquirerAPIKey  string `mapstructure:"ACQUIRER_API_KEY"`

	// Admin API key gating the /v1/admin operator console. Presented via the
	// X-Admin-Key header and compared with a constant-time compare. Empty
	// disables the admin surface entirely (every /v1/admin request 401s).
	AdminAPIKey string `mapstructure:"ADMIN_API_KEY"`

	// WebDir is the directory served as static files so the dashboard / admin /
	// checkout UIs run same-origin with the API (avoiding CORS). Empty disables
	// static serving.
	WebDir string `mapstructure:"WEB_DIR"`

	// Shared secret used to verify inbound QR payment webhooks (HMAC-SHA256).
	QRWebhookSecret string `mapstructure:"QR_WEBHOOK_SECRET"`

	// Shared secret used to sign OUTBOUND webhook deliveries (HMAC-SHA256 over
	// the raw JSON body, sent in X-Signature). Merchants verify against this.
	WebhookSigningSecret string `mapstructure:"WEBHOOK_SIGNING_SECRET"`

	// Fallback endpoint the webhook delivery worker POSTs to when an event row
	// carries no per-merchant target_url (dev/sandbox convenience).
	WebhookDefaultURL string `mapstructure:"WEBHOOK_DEFAULT_URL"`

	// Max delivery attempts before an outbound webhook is parked as failed.
	WebhookMaxAttempts int `mapstructure:"WEBHOOK_MAX_ATTEMPTS"`

	// Settlement fee charged on captured volume, in basis points (250 = 2.50%).
	SettlementFeeBps int64 `mapstructure:"SETTLEMENT_FEE_BPS"`

	// Origin used to build simulated QR image URLs (mock provider). Empty falls
	// back to the provider's sandbox default.
	QRProviderBaseURL string `mapstructure:"QR_PROVIDER_BASE_URL"`

	// SandboxMode, when true, mounts the public sandbox payer-simulator endpoints
	// (/v1/sandbox/...) that let a browser mark a QR payment paid/failed WITHOUT a
	// merchant API key (the server signs the confirmation internally). It MUST stay
	// false in any real deploy: when false every sandbox route is completely absent
	// (404), so a production gateway can never let someone mark payments paid.
	SandboxMode bool `mapstructure:"SANDBOX_MODE"`

	// Default PromptPay proxy the QR funds settle to. In production this is
	// resolved per-merchant from the merchant profile; here it is injected via
	// config so locally-built PromptPay QR has a target. Set exactly one.
	PromptPayMobileNo   string `mapstructure:"PROMPTPAY_MOBILE_NO"`
	PromptPayNationalID string `mapstructure:"PROMPTPAY_NATIONAL_ID"`
	PromptPayEWalletID  string `mapstructure:"PROMPTPAY_EWALLET_ID"`
}

// envKeys is every config env var Unmarshal must read. It must stay in sync with
// the mapstructure tags on Config. Keys that already have a SetDefault are still
// listed here for completeness (binding them is harmless); the ones that matter
// are the secret / DSN / TLS keys with no default, which otherwise never
// unmarshal from a pure-env (no .env file) environment such as Railway.
var envKeys = []string{
	"ENV", "HTTP_ADDR", "LOG_LEVEL",
	"TLS_CERT_FILE", "TLS_KEY_FILE", "TLS_TERMINATED_UPSTREAM",
	"CORS_ALLOW_ORIGINS", "RATE_LIMIT_PER_SEC", "SIGNUP_RATE_LIMIT_PER_HOUR",
	"MIGRATE_ON_BOOT", "DB_MAX_CONNS", "DB_MIN_CONNS",
	"METRICS_ADDR", "METRICS_PUBLIC", "DATABASE_URL",
	"JWT_SECRET", "IDEMPOTENCY_TTL_SECONDS", "KMS_KEY_ID", "VAULT_NAMESPACE",
	"ACQUIRER_BASE_URL", "ACQUIRER_API_KEY", "ADMIN_API_KEY", "WEB_DIR",
	"QR_WEBHOOK_SECRET", "WEBHOOK_SIGNING_SECRET", "WEBHOOK_DEFAULT_URL",
	"WEBHOOK_MAX_ATTEMPTS", "SETTLEMENT_FEE_BPS", "QR_PROVIDER_BASE_URL",
	"PROMPTPAY_MOBILE_NO", "PROMPTPAY_NATIONAL_ID", "PROMPTPAY_EWALLET_ID",
	"SANDBOX_MODE",
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetDefault("ENV", "development")
	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("IDEMPOTENCY_TTL_SECONDS", 86400)
	v.SetDefault("WEBHOOK_MAX_ATTEMPTS", 6)
	v.SetDefault("SETTLEMENT_FEE_BPS", 0)
	v.SetDefault("RATE_LIMIT_PER_SEC", 600)
	v.SetDefault("SIGNUP_RATE_LIMIT_PER_HOUR", 5)
	v.SetDefault("MIGRATE_ON_BOOT", false)
	v.SetDefault("DB_MAX_CONNS", 50)
	v.SetDefault("DB_MIN_CONNS", 10)
	v.SetDefault("METRICS_PUBLIC", false)
	v.SetDefault("WEB_DIR", "./web")
	v.SetDefault("SANDBOX_MODE", false)

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // .env is optional; env vars win.

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind every config key to its environment variable. viper's
	// AutomaticEnv makes env vars available to Get(), but Unmarshal only maps keys
	// it already knows about — those with a SetDefault or that appeared in a read
	// config file. On a PaaS like Railway there is NO .env file and secrets arrive
	// purely as env vars, so without these binds keys like DATABASE_URL / the
	// production secrets would silently unmarshal empty and the prod fail-fast
	// would reject a correctly-configured deploy. Binding registers them so
	// Unmarshal reads them from the environment regardless of a .env file.
	for _, key := range envKeys {
		_ = v.BindEnv(key)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	// PORT override (PaaS convention: Railway/Heroku/Fly inject $PORT). viper's
	// Unmarshal does not auto-bind PORT (it maps to no struct field), so read it
	// explicitly. When set it wins over HTTP_ADDR and binds on all interfaces
	// (0.0.0.0) so the platform edge can route to the container. HTTP_ADDR stays
	// the local default (:8080) when PORT is unset.
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		c.HTTPAddr = "0.0.0.0:" + strings.TrimPrefix(port, ":")
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// validate fails fast on missing required config. Card-data keys (KMS/vault) are
// only mandatory once the vault is wired; here we enforce the operational DB and
// strong secrets in production so the service never starts up half-connected or
// with guessable webhook/JWT secrets that would let an attacker forge callbacks.
func (c *Config) validate() error {
	// Normalize the webhook retry budget into a sane, small range for every env.
	// This also guarantees the value fits an int32 when handed to the worker
	// (which types the attempts column as int32), so the downstream conversion
	// can never overflow.
	if c.WebhookMaxAttempts < 1 {
		c.WebhookMaxAttempts = 1
	}
	if c.WebhookMaxAttempts > 100 {
		c.WebhookMaxAttempts = 100
	}

	if !c.IsProd() {
		return nil
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("config: DATABASE_URL is required when ENV=production")
	}
	// Require strong, non-default shared secrets in production. A guessable
	// QR webhook secret lets an attacker forge 'paid' confirmations; a weak
	// outbound signing secret lets them forge merchant deliveries.
	required := map[string]string{
		"QR_WEBHOOK_SECRET":      c.QRWebhookSecret,
		"WEBHOOK_SIGNING_SECRET": c.WebhookSigningSecret,
		"JWT_SECRET":             c.JWTSecret,
	}
	for name, val := range required {
		if err := requireStrongSecret(name, val); err != nil {
			return err
		}
	}
	// Enforce transport security in production. Refuse to start over plaintext
	// HTTP unless TLS is configured on this process OR the operator explicitly
	// acknowledges that a trusted upstream terminates TLS. A card-acquiring
	// gateway must never silently serve cleartext in prod.
	if !c.TLSEnabled() && !c.TLSTerminatedUpstream {
		return fmt.Errorf("config: refusing to start ENV=production over plaintext HTTP: set TLS_CERT_FILE+TLS_KEY_FILE, or set TLS_TERMINATED_UPSTREAM=true to acknowledge upstream TLS termination")
	}
	return nil
}

// requireStrongSecret rejects empty, placeholder, or short secrets in production.
func requireStrongSecret(name, val string) error {
	v := strings.TrimSpace(val)
	if v == "" {
		return fmt.Errorf("config: %s is required when ENV=production", name)
	}
	if strings.EqualFold(v, "change-me-in-prod") {
		return fmt.Errorf("config: %s is set to the placeholder default; set a real secret when ENV=production", name)
	}
	if len(v) < 32 {
		return fmt.Errorf("config: %s must be at least 32 bytes when ENV=production", name)
	}
	return nil
}

// CORSAllowOriginList returns the configured CORS allowlist as a normalized,
// comma-separated string suitable for the fiber cors middleware. Empty when no
// origins are configured (CORS disabled — the safe default for a money API).
func (c *Config) CORSAllowOriginList() string {
	parts := strings.Split(c.CORSAllowOrigins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" && p != "*" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}

func (c *Config) LogLevel() zerolog.Level {
	lvl, err := zerolog.ParseLevel(c.LogLevelStr)
	if err != nil {
		return zerolog.InfoLevel
	}
	return lvl
}

func (c *Config) IsProd() bool { return strings.EqualFold(c.Env, "production") }

// IsSandbox reports whether the public sandbox payer-simulator endpoints should
// be mounted. It is the single gate the router checks; when false the sandbox
// routes are completely absent (404) so a real deploy can never mark payments
// paid without a signed bank webhook.
func (c *Config) IsSandbox() bool { return c.SandboxMode }

// TLSEnabled reports whether both a certificate and key file are configured, in
// which case the server should listen over TLS.
func (c *Config) TLSEnabled() bool {
	return strings.TrimSpace(c.TLSCertFile) != "" && strings.TrimSpace(c.TLSKeyFile) != ""
}
