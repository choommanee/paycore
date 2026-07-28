package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// SecretBox seals and opens arbitrary application secrets (e.g. a per-merchant
// webhook signing secret) using the SAME envelope-encryption scheme and KMS as
// the card-data Vault: a fresh per-record data-encryption key (DEK) encrypts the
// plaintext with AES-256-GCM, and the DEK is wrapped by the KMS.
//
// It exists as a distinct type from Vault because Vault.Tokenize deliberately
// validates PAN shape (12–19 digits) and returns a base64 token + last4 — the
// wrong contract for a 70-char "whsec_…" secret. SecretBox is content-agnostic
// and returns raw bytes suitable for a BYTEA column.
//
// SECURITY: the sealed blob is safe at rest (only the KMS can unwrap the DEK).
// Never log the plaintext secret or the DEK. Open only where the secret is
// immediately needed (e.g. signing an outbound delivery) and discard it after.
type SecretBox struct {
	kms KMS
}

// NewSecretBox constructs a SecretBox backed by the given KMS.
func NewSecretBox(kms KMS) *SecretBox { return &SecretBox{kms: kms} }

// secretBoxV1 versions the sealed-blob layout so it can evolve unambiguously.
//
//	[1]byte  version (0x01)
//	[2]byte  big-endian uint16 len(wrappedDEK)
//	[N]byte  wrappedDEK
//	[..]     nonce||ciphertext (AES-256-GCM of the plaintext under the DEK)
const secretBoxV1 = 0x01

// Seal envelope-encrypts plaintext and returns an opaque, self-describing blob.
func (s *SecretBox) Seal(ctx context.Context, plaintext []byte) ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("crypto: dek: %w", err)
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: gcm: %w", err)
	}
	ct, err := sealGCM(gcm, plaintext)
	if err != nil {
		return nil, err
	}
	wrappedDEK, err := s.kms.WrapDEK(ctx, dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap dek: %w", err)
	}
	if len(wrappedDEK) > 0xFFFF {
		return nil, fmt.Errorf("crypto: wrapped dek too large: %d bytes", len(wrappedDEK))
	}

	out := make([]byte, 0, 1+2+len(wrappedDEK)+len(ct))
	out = append(out, secretBoxV1)
	var lenb [2]byte
	binary.BigEndian.PutUint16(lenb[:], uint16(len(wrappedDEK))) // #nosec G115 -- bounded above
	out = append(out, lenb[:]...)
	out = append(out, wrappedDEK...)
	out = append(out, ct...)
	return out, nil
}

// Open reverses Seal, returning the plaintext secret.
func (s *SecretBox) Open(ctx context.Context, blob []byte) ([]byte, error) {
	if len(blob) < 3 || blob[0] != secretBoxV1 {
		return nil, fmt.Errorf("crypto: unrecognized secret blob")
	}
	dekLen := int(binary.BigEndian.Uint16(blob[1:3]))
	if len(blob) < 3+dekLen {
		return nil, fmt.Errorf("crypto: secret blob truncated")
	}
	wrappedDEK := blob[3 : 3+dekLen]
	ct := blob[3+dekLen:]

	dek, err := s.kms.UnwrapDEK(ctx, wrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("crypto: unwrap dek: %w", err)
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: gcm: %w", err)
	}
	pt, err := openGCM(gcm, ct)
	if err != nil {
		return nil, err
	}
	return pt, nil
}
