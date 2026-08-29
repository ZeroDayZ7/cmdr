// internal/crypto/argon2.go
package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Argon2Params definiuje parametry derywacji klucza
type Argon2Params struct {
	Time    uint32 // liczba iteracji (m_cost / t_cost)
	Memory  uint32 // pamięć w KiB (np. 64 * 1024 = 64MB)
	Threads uint8  // liczba wątków
	KeyLen  uint32 // długość wygenerowanego klucza w bajtach (32 dla AES-256)
	SaltLen uint32 // długość soli w bajtach
}

// DefaultArgon2Params zwraca domyślne, bezpieczne parametry OWASP
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Time:    3,
		Memory:  64 * 1024, // 64 MB
		Threads: 4,
		KeyLen:  32, // 256 bitów dla AES-256
		SaltLen: 16, // 16 bajtów soli
	}
}

// DeriveKeyArgon2id generuje klucz AES (32 bajty) oraz nową sól na podstawie hasła
func DeriveKeyArgon2id(password string, params Argon2Params) (key []byte, salt []byte, err error) {
	salt = make([]byte, params.SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	key = argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)
	return key, salt, nil
}

// DeriveKeyWithSaltArgon2id generuje klucz AES na podstawie podanego hasła i istniejącej soli (używane przy odszyfrowywaniu)
func DeriveKeyWithSaltArgon2id(password string, salt []byte, params Argon2Params) []byte {
	return argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)
}
