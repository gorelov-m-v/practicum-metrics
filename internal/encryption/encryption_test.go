package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	plaintext := []byte("test payload")

	ciphertext, err := Encrypt(plaintext, &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, privateKey)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestLoadKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	dir := t.TempDir()

	privateKeyPath := filepath.Join(dir, "private.pem")
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	publicKeyPath := filepath.Join(dir, "public.pem")
	if err := os.WriteFile(publicKeyPath, pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}), 0o600); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	loadedPublicKey, err := LoadPublicKey(publicKeyPath)
	if err != nil {
		t.Fatalf("load public key: %v", err)
	}

	loadedPrivateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		t.Fatalf("load private key: %v", err)
	}

	ciphertext, err := Encrypt([]byte("hello"), loadedPublicKey)
	if err != nil {
		t.Fatalf("encrypt with loaded public key: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, loadedPrivateKey)
	if err != nil {
		t.Fatalf("decrypt with loaded private key: %v", err)
	}

	if string(decrypted) != "hello" {
		t.Fatalf("expected decrypted payload to match original")
	}
}
