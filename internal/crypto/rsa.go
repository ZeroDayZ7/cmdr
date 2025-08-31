package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// GenerateRSAKey generuje parę kluczy RSA i zapisuje je w plikach PEM
func GenerateRSAKey(bits int, privatePath, publicPath string) error {
	if bits != 2048 && bits != 3072 && bits != 4096 {
		return fmt.Errorf("niewspierany rozmiar klucza: %d (użyj 2048, 3072 lub 4096)", bits)
	}

	// Generowanie klucza prywatnego
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("błąd generowania klucza: %w", err)
	}

	// Kodowanie klucza prywatnego do PEM
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	if err := os.WriteFile(privatePath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("błąd zapisu klucza prywatnego: %w", err)
	}

	// Generowanie klucza publicznego
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("błąd serializacji klucza publicznego: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)

	if err := os.WriteFile(publicPath, publicKeyPEM, 0644); err != nil {
		return fmt.Errorf("błąd zapisu klucza publicznego: %w", err)
	}

	return nil
}
