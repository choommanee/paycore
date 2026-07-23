package config

import (
	"strings"
	"testing"
)

const strongSecret = "a-sufficiently-long-production-secret-32b"

func prodConfig() *Config {
	return &Config{
		Env:                   "production",
		DatabaseURL:           "postgres://u:p@h:5432/db",
		QRWebhookSecret:       strongSecret,
		WebhookSigningSecret:  strongSecret,
		JWTSecret:             strongSecret,
		TLSTerminatedUpstream: true, // acknowledge upstream TLS termination
	}
}

func TestValidateProdRequiresStrongSecrets(t *testing.T) {
	// Baseline: a fully-configured prod config validates.
	if err := prodConfig().validate(); err != nil {
		t.Fatalf("valid prod config rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{"missing db", func(c *Config) { c.DatabaseURL = "" }, "DATABASE_URL"},
		{"placeholder qr secret", func(c *Config) { c.QRWebhookSecret = "change-me-in-prod" }, "QR_WEBHOOK_SECRET"},
		{"empty signing secret", func(c *Config) { c.WebhookSigningSecret = "" }, "WEBHOOK_SIGNING_SECRET"},
		{"short jwt secret", func(c *Config) { c.JWTSecret = "too-short" }, "JWT_SECRET"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := prodConfig()
			tc.mutate(c)
			err := c.validate()
			if err == nil {
				t.Fatalf("expected validation error mentioning %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestValidateProdRequiresTransportSecurity(t *testing.T) {
	// Neither TLS configured nor upstream-termination acknowledged: must refuse.
	c := prodConfig()
	c.TLSTerminatedUpstream = false
	err := c.validate()
	if err == nil {
		t.Fatal("expected prod config over plaintext HTTP to be rejected")
	}
	if !strings.Contains(err.Error(), "plaintext HTTP") {
		t.Fatalf("error %q does not mention plaintext HTTP transport", err)
	}

	// TLS configured on the process: valid without the ack flag.
	c2 := prodConfig()
	c2.TLSTerminatedUpstream = false
	c2.TLSCertFile = "/etc/tls/cert.pem"
	c2.TLSKeyFile = "/etc/tls/key.pem"
	if err := c2.validate(); err != nil {
		t.Fatalf("prod config with TLS files rejected: %v", err)
	}

	// Explicit upstream-termination ack: valid without cert files.
	c3 := prodConfig() // ack already true
	if err := c3.validate(); err != nil {
		t.Fatalf("prod config with upstream TLS ack rejected: %v", err)
	}
}

func TestValidateDevSkipsSecretChecks(t *testing.T) {
	// A dev config with placeholder secrets and no DB must still validate.
	c := &Config{Env: "development"}
	if err := c.validate(); err != nil {
		t.Fatalf("dev config rejected: %v", err)
	}
}

func TestLoadPortOverridesHTTPAddr(t *testing.T) {
	// PaaS convention: when $PORT is set it wins over HTTP_ADDR and binds on all
	// interfaces (0.0.0.0) so the platform edge can reach the container.
	t.Setenv("PORT", "9137")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() with PORT set failed: %v", err)
	}
	if c.HTTPAddr != "0.0.0.0:9137" {
		t.Fatalf("HTTPAddr=%q, want 0.0.0.0:9137", c.HTTPAddr)
	}
}

func TestLoadPortAcceptsColonPrefix(t *testing.T) {
	// A ':9137' style value is normalized to a single 0.0.0.0:9137 bind address.
	t.Setenv("PORT", ":9137")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() with PORT=:9137 failed: %v", err)
	}
	if c.HTTPAddr != "0.0.0.0:9137" {
		t.Fatalf("HTTPAddr=%q, want 0.0.0.0:9137", c.HTTPAddr)
	}
}

func TestLoadDefaultsWhenPortUnset(t *testing.T) {
	// Without $PORT, HTTP_ADDR keeps its local default and the new knobs default.
	t.Setenv("PORT", "")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if c.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr=%q, want :8080", c.HTTPAddr)
	}
	if c.SignupRateLimitPerHour != 5 {
		t.Fatalf("SignupRateLimitPerHour=%d, want 5", c.SignupRateLimitPerHour)
	}
	if c.MigrateOnBoot {
		t.Fatal("MigrateOnBoot should default to false")
	}
}

func TestCORSAllowOriginListDropsWildcardAndBlanks(t *testing.T) {
	c := &Config{CORSAllowOrigins: "https://a.example.com, , * ,https://b.example.com"}
	got := c.CORSAllowOriginList()
	want := "https://a.example.com,https://b.example.com"
	if got != want {
		t.Fatalf("CORSAllowOriginList()=%q want %q", got, want)
	}
}
