package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Vault tokenizes card PANs using envelope encryption. Each Tokenize call mints
// a fresh per-record data-encryption key (DEK), encrypts the PAN with it
// (AES-256-GCM), and wraps the DEK via the KMS. The resulting token embeds both
// the wrapped DEK and the encrypted PAN, so Detokenize is fully reversible with
// only the token plus KMS access.
//
// Storage model: callers persist ONLY the token and the last4. The full PAN is
// never written in plaintext anywhere. In a production PCI environment the
// token would additionally be stored in a segmented vault database.
//
// SECURITY: never log the PAN, the DEK, or the KMS master key. Detokenize should
// only be called from inside the PCI scope (e.g. right before an acquirer
// authorization) and the plaintext PAN must be discarded immediately after.
type Vault struct {
	kms KMS
}

// tokenPrefix versions the token format so it can evolve without ambiguity.
const tokenPrefix = "tokv1"

// NewVault constructs a Vault backed by the given KMS.
func NewVault(kms KMS) *Vault {
	return &Vault{kms: kms}
}

// Tokenize encrypts pan and returns an opaque token plus the last 4 digits.
// The plaintext PAN is not retained by the vault.
func (v *Vault) Tokenize(ctx context.Context, pan string) (token string, last4 string, err error) {
	pan = strings.TrimSpace(pan)
	if len(pan) < 12 || len(pan) > 19 {
		return "", "", fmt.Errorf("crypto: PAN length out of range")
	}
	last4 = pan[len(pan)-4:]

	// Fresh per-record DEK.
	dek := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, dek); err != nil {
		return "", "", fmt.Errorf("crypto: dek: %w", err)
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", "", fmt.Errorf("crypto: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("crypto: gcm: %w", err)
	}
	encPAN, err := sealGCM(gcm, []byte(pan))
	if err != nil {
		return "", "", err
	}

	wrappedDEK, err := v.kms.WrapDEK(ctx, dek)
	if err != nil {
		return "", "", fmt.Errorf("crypto: wrap dek: %w", err)
	}

	token, err = encodeToken(wrappedDEK, encPAN)
	if err != nil {
		return "", "", err
	}
	return token, last4, nil
}

// Detokenize reverses Tokenize, returning the plaintext PAN. Callers must handle
// the returned PAN inside PCI scope and discard it immediately.
func (v *Vault) Detokenize(ctx context.Context, token string) (string, error) {
	wrappedDEK, encPAN, err := decodeToken(token)
	if err != nil {
		return "", err
	}
	dek, err := v.kms.UnwrapDEK(ctx, wrappedDEK)
	if err != nil {
		return "", fmt.Errorf("crypto: unwrap dek: %w", err)
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", fmt.Errorf("crypto: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: gcm: %w", err)
	}
	pan, err := openGCM(gcm, encPAN)
	if err != nil {
		return "", err
	}
	return string(pan), nil
}

// encodeToken serializes the wrapped DEK and encrypted PAN into a single
// URL-safe base64 string with a version prefix. Layout:
//
//	"tokv1." + base64( uint16(len(wrappedDEK)) || wrappedDEK || encPAN )
func encodeToken(wrappedDEK, encPAN []byte) (string, error) {
	// The length prefix is a uint16, so a wrapped DEK must fit in 2 bytes. A KMS
	// wrap of a 256-bit DEK is on the order of tens of bytes, so this bound is
	// never hit in practice — but reject explicitly rather than silently
	// truncating the length header (which would corrupt the token).
	if len(wrappedDEK) > 0xFFFF {
		return "", fmt.Errorf("crypto: wrapped dek too large: %d bytes", len(wrappedDEK))
	}
	buf := make([]byte, 2+len(wrappedDEK)+len(encPAN))
	binary.BigEndian.PutUint16(buf[:2], uint16(len(wrappedDEK))) // #nosec G115 -- length bounded to uint16 range by the check above
	copy(buf[2:], wrappedDEK)
	copy(buf[2+len(wrappedDEK):], encPAN)
	return tokenPrefix + "." + base64.RawURLEncoding.EncodeToString(buf), nil
}

// decodeToken reverses encodeToken.
func decodeToken(token string) (wrappedDEK, encPAN []byte, err error) {
	prefix, b64, ok := strings.Cut(token, ".")
	if !ok || prefix != tokenPrefix {
		return nil, nil, fmt.Errorf("crypto: unrecognized token format")
	}
	buf, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return nil, nil, fmt.Errorf("crypto: token base64: %w", err)
	}
	if len(buf) < 2 {
		return nil, nil, fmt.Errorf("crypto: token truncated")
	}
	dekLen := int(binary.BigEndian.Uint16(buf[:2]))
	if len(buf) < 2+dekLen {
		return nil, nil, fmt.Errorf("crypto: token truncated")
	}
	wrappedDEK = buf[2 : 2+dekLen]
	encPAN = buf[2+dekLen:]
	return wrappedDEK, encPAN, nil
}
