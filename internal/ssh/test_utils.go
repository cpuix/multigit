package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfig holds configuration for SSH tests
type TestConfig struct {
	TempDir    string
	KeyFiles   []string
	ConfigFile string
}

// SetupTestEnvironment creates a test environment with temporary files
func SetupTestEnvironment(t *testing.T) *TestConfig {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	configFile := filepath.Join(tempDir, "config")
	if _, err := os.Create(configFile); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create config file: %v", err)
	}

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Test failed, preserving temp dir: %s", tempDir)
			return
		}
		os.RemoveAll(tempDir)
	})

	return &TestConfig{
		TempDir:    tempDir,
		ConfigFile: configFile,
	}
}

// CreateTestKey creates a test SSH key pair with proper permissions
func (tc *TestConfig) CreateTestKey(t *testing.T, keyName string, keyType KeyType) string {
	t.Helper()

	keyFile := filepath.Join(tc.TempDir, keyName)

	switch keyType {
	case KeyTypeRSA:
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("Failed to generate RSA key: %v", err)
		}

		// Write private key with restricted permissions (0600)
		privKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		}

		// Create the file with restricted permissions
		privKeyFile, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			t.Fatalf("Failed to create private key file: %v", err)
		}
		defer privKeyFile.Close()

		if err := pem.Encode(privKeyFile, privKeyPEM); err != nil {
			t.Fatalf("Failed to write private key: %v", err)
		}

		// Write public key
		pubKey, err := sshPublicKeyRSA(&privKey.PublicKey, "test@example.com")
		if err != nil {
			t.Fatalf("Failed to generate public key: %v", err)
		}

		if err := os.WriteFile(keyFile+".pub", pubKey, 0644); err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}

	case KeyTypeED25519:
		// For simplicity, we'll create empty files for ED25519 in tests
		// with proper permissions
		privFile, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			t.Fatalf("Failed to create ED25519 key file: %v", err)
		}
		privFile.Close()

		pubFile, err := os.Create(keyFile + ".pub")
		if err != nil {
			t.Fatalf("Failed to create ED25519 pubkey file: %v", err)
		}
		pubFile.Close()
	}

	tc.KeyFiles = append(tc.KeyFiles, keyFile)
	return keyFile
}

// AssertFileExists checks if a file exists and fails the test if it doesn't
func (tc *TestConfig) AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks if a file doesn't exist and fails the test if it does
func (tc *TestConfig) AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Expected file to not exist: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Error checking file existence: %v", err)
	}
}

// AssertConfigContains checks if the SSH config contains the given string
func (tc *TestConfig) AssertConfigContains(t *testing.T, contains string) {
	t.Helper()
	configData, err := os.ReadFile(tc.ConfigFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(configData), contains) {
		t.Fatalf("Expected config to contain %q, got: %s", contains, string(configData))
	}
}

// AssertConfigNotContains checks if the SSH config doesn't contain the given string
func (tc *TestConfig) AssertConfigNotContains(t *testing.T, contains string) {
	t.Helper()
	configData, err := os.ReadFile(tc.ConfigFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if strings.Contains(string(configData), contains) {
		t.Fatalf("Expected config to not contain %q, got: %s", contains, string(configData))
	}
}
