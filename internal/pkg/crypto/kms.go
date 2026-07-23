// Package crypto provides the card-data tokenization vault used by the payment
// gateway. It implements envelope encryption: a data-encryption key (DEK)
// encrypts the PAN, and a key-management service (KMS) encrypts the DEK.
//
// The KMS interface abstracts the key custodian so the real deployment can plug
// in AWS KMS, GCP KMS, HashiCorp Vault Transit, or a network HSM. The bundled
// LocalKMS is a DEVELOPMENT-ONLY stand-in that keeps the master key in process
// memory.
//
// PRODUCTION WARNING: never use LocalKMS in production. A real HSM/KMS must hold
// the master key so plaintext key material never leaves the security boundary,
// and PCI-DSS key rotation / dual control can be enforced. This package only
// ever persists a token + last4 conceptually; the full PAN is never stored in
// plaintext.
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KMS is the key-management boundary. Implementations wrap (encrypt) and unwrap
// (decrypt) data-encryption keys using a master key that never leaves the KMS.
//
// Ciphertext returned by WrapDEK is opaque to callers and must round-trip
// through the same KMS (and key) to be unwrapped.
type KMS interface {
	// WrapDEK encrypts a plaintext data-encryption key.
	WrapDEK(ctx context.Context, plaintextDEK []byte) (wrappedDEK []byte, err error)
	// UnwrapDEK decrypts a previously wrapped data-encryption key.
	UnwrapDEK(ctx context.Context, wrappedDEK []byte) (plaintextDEK []byte, err error)
	// KeyID identifies the master key, for audit and rotation bookkeeping.
	KeyID() string
}

// LocalKMS is an in-process, AES-256-GCM KMS intended ONLY for local
// development and tests. The 32-byte master key lives in memory.
//
// DO NOT USE IN PRODUCTION: production must use a real HSM/KMS (AWS KMS, GCP
// KMS, Vault Transit, PKCS#11 HSM) so the master key is non-exportable.
type LocalKMS struct {
	keyID string
	gcm   cipher.AEAD
}

// compile-time assertion.
var _ KMS = (*LocalKMS)(nil)

// NewLocalKMS builds a LocalKMS from a 32-byte master key (AES-256). The keyID
// is recorded for audit parity with a real KMS alias.
func NewLocalKMS(keyID string, masterKey []byte) (*LocalKMS, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("crypto: local KMS master key must be 32 bytes, got %d", len(masterKey))
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: local KMS cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: local KMS gcm: %w", err)
	}
	return &LocalKMS{keyID: keyID, gcm: gcm}, nil
}

// NewDevLocalKMS returns a LocalKMS seeded with a fixed, well-known DEV key so
// tokens are reproducible across restarts in development. This key is public and
// worthless; never rely on it for anything real.
func NewDevLocalKMS(keyID string) (*LocalKMS, error) {
	// 32 bytes of an obvious dev pattern.
	devKey := []byte("dev-only-cardvault-master-key!!32")
	return NewLocalKMS(keyID, devKey[:32])
}

// KeyID returns the master key identifier.
func (k *LocalKMS) KeyID() string { return k.keyID }

// WrapDEK encrypts the DEK with AES-256-GCM, prefixing a random nonce.
func (k *LocalKMS) WrapDEK(_ context.Context, plaintextDEK []byte) ([]byte, error) {
	return sealGCM(k.gcm, plaintextDEK)
}

// UnwrapDEK reverses WrapDEK.
func (k *LocalKMS) UnwrapDEK(_ context.Context, wrappedDEK []byte) ([]byte, error) {
	return openGCM(k.gcm, wrappedDEK)
}

// sealGCM encrypts plaintext with the given AEAD, returning nonce||ciphertext.
func sealGCM(gcm cipher.AEAD, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// openGCM reverses sealGCM.
func openGCM(gcm cipher.AEAD, blob []byte) ([]byte, error) {
	ns := gcm.NonceSize()
	if len(blob) < ns {
		return nil, fmt.Errorf("crypto: ciphertext too short")
	}
	nonce, ct := blob[:ns], blob[ns:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt: %w", err)
	}
	return pt, nil
}
