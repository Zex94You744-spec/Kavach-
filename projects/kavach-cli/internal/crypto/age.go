// internal/crypto/age.go
package crypto

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

func Encrypt(plaintext []byte, passphrase []byte) ([]byte, error) {
	// age API string expect karta hai, short-lived copy banti hai (acceptable)
	recipient, err := age.NewScryptRecipient(string(passphrase))
	if err != nil {
		return nil, fmt.Errorf("failed to create scrypt recipient: %w", err)
	}

	buf := new(bytes.Buffer)
	w, err := age.Encrypt(buf, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	if _, err := io.Copy(w, bytes.NewReader(plaintext)); err != nil {
		return nil, fmt.Errorf("failed to write plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encryptor: %w", err)
	}

	return buf.Bytes(), nil
}

func Decrypt(ciphertext []byte, passphrase []byte) ([]byte, error) {
	identity, err := age.NewScryptIdentity(string(passphrase))
	if err != nil {
		return nil, fmt.Errorf("failed to create scrypt identity: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, fmt.Errorf("failed to read decrypted  %w", err)
	}

	return out.Bytes(), nil
}
