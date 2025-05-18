//go:build integration

package ssh_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSHKeyLifecycle, bir SSH anahtarının tüm yaşam döngüsünü test eder:
// 1. Yeni bir SSH anahtarı oluşturma
// 2. SSH agent'a ekleme
// 3. SSH config'e giriş ekleme
// 4. Anahtarı silme
func TestSSHKeyLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Geçici bir .ssh dizini oluştur
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	err := os.Mkdir(sshDir, 0700)
	require.NoError(t, err, "Failed to create .ssh directory")

	// Geçici SSH config dosyası oluştur
	configPath := filepath.Join(sshDir, "config")
	_, err = os.Create(configPath)
	require.NoError(t, err, "Failed to create SSH config file")

	// Test verilerini hazırla
	accountName := "integration-test-account" + t.Name()
	email := "integration-test@example.com"
	keyFile := filepath.Join(sshDir, "id_ed25519_"+accountName)

	// Orijinal home değerini sakla ve test sonunda geri yükle
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	t.Run("Create SSH key", func(t *testing.T) {
		err := ssh.CreateSSHKey(accountName, email, keyFile, ssh.KeyTypeED25519)
		require.NoError(t, err, "SSH anahtarı oluşturulamadı")

		// Anahtar dosyalarının oluştuğunu doğrula
		_, err = os.Stat(keyFile)
		assert.NoError(t, err, "Özel anahtar dosyası oluşturulmadı")
		_, err = os.Stat(keyFile + ".pub")
		assert.NoError(t, err, "Genel anahtar dosyası oluşturulmadı")
	})

	t.Run("Add key to SSH agent", func(t *testing.T) {
		// SSH agent'ı başlat (eğer çalışmıyorsa)
		if os.Getenv("SSH_AUTH_SOCK") == "" {
			t.Skip("SSH agent çalışmıyor, test atlanıyor")
		}

		// Anahtarın tam yolunu kullanarak ekle
		err := ssh.AddSSHKeyToAgent(keyFile)
		assert.NoError(t, err, "SSH anahtarı agent'a eklenemedi")

		// SSH agent'ta anahtarın olduğunu doğrula
		cmd := exec.Command("ssh-add", "-l")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "ssh-add -l komutu başarısız oldu")
		// SSH agent çıktısında anahtarın parmak izini kontrol et
		assert.Contains(t, string(output), filepath.Base(keyFile), "SSH agent'ta anahtar bulunamadı")
	})

	t.Run("Add SSH config entry", func(t *testing.T) {
		// SSH config dosyasına giriş ekle
		err := ssh.AddSSHConfigEntry(accountName)
		require.NoError(t, err, "SSH config girişi eklenemedi")

		// Config dosyasının güncellendiğini doğrula
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err, "SSH config dosyası okunamadı")

		// SSH config dosyasında beklenen girişlerin olduğunu kontrol et
		expectedHost := fmt.Sprintf("github.com-%s", accountName)
		configContent := string(configData)

		assert.Contains(t, configContent, "Host "+expectedHost,
			"SSH config dosyasında beklenen Host bulunamadı")
		assert.Contains(t, configContent, "IdentityFile "+keyFile,
			"SSH config dosyasında IdentityFile bulunamadı")
	})

	t.Run("Delete SSH key", func(t *testing.T) {
		err := ssh.DeleteSSHKey(accountName)
		assert.NoError(t, err, "SSH anahtarı silinemedi")

		// Anahtar dosyalarının silindiğini doğrula
		_, err = os.Stat(keyFile)
		assert.True(t, os.IsNotExist(err), "Özel anahtar dosyası silinmedi")
		_, err = os.Stat(keyFile + ".pub")
		assert.True(t, os.IsNotExist(err), "Genel anahtar dosyası silinmedi")

		// SSH agent'tan kaldırıldığını doğrula
		if os.Getenv("SSH_AUTH_SOCK") != "" {
			cmd := exec.Command("ssh-add", "-l")
			output, err := cmd.CombinedOutput()
			if err == nil { // Eğer hata yoksa (yani liste boş değilse)
				assert.NotContains(t, string(output), filepath.Base(keyFile),
					"SSH agent'tan anahtar kaldırılmadı")
			}
		}

		// SSH config'ten kaldırıldığını doğrula
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err, "SSH config dosyası okunamadı")
		assert.NotContains(t, string(configData), "Host github.com-"+accountName,
			"SSH config'ten anahtar kaldırılmadı")
	})
}

// TestSSHKeyFileOperations, doğrudan dosya işlemlerini test eder
func TestSSHKeyFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Geçici bir .ssh dizini oluştur
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	err := os.Mkdir(sshDir, 0700)
	require.NoError(t, err, "Failed to create .ssh directory")

	// Orijinal home değerini sakla ve test sonunda geri yükle
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	keyFile := filepath.Join(sshDir, "test_key")

	t.Run("Create and delete key file directly", func(t *testing.T) {
		// Anahtar oluştur
		err := ssh.CreateSSHKey("file-test", "test@example.com", keyFile, ssh.KeyTypeED25519)
		require.NoError(t, err, "SSH anahtarı oluşturulamadı")

		// Doğrudan dosya silme işlemini test et
		err = ssh.DeleteSSHKeyFile(keyFile)
		assert.NoError(t, err, "SSH anahtar dosyaları silinemedi")

		// Dosyaların silindiğini doğrula
		_, err = os.Stat(keyFile)
		assert.True(t, os.IsNotExist(err), "Özel anahtar dosyası silinmedi")
		_, err = os.Stat(keyFile + ".pub")
		assert.True(t, os.IsNotExist(err), "Genel anahtar dosyası silinmedi")
	})
}
