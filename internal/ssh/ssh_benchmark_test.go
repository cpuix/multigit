package ssh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuix/multigit/internal/ssh"
)

func BenchmarkCreateSSHKey(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "multigit-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		keyFile := filepath.Join(tempDir, "test_key")
		err := ssh.CreateSSHKey("benchmark", "benchmark@example.com", keyFile, ssh.KeyTypeRSA)
		if err != nil {
			b.Fatalf("Failed to create SSH key: %v", err)
		}
	}
}

func BenchmarkAddSSHKeyToAgent(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "multigit-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	keyFile := filepath.Join(tempDir, "test_key")
	err = ssh.CreateSSHKey("benchmark", "benchmark@example.com", keyFile, ssh.KeyTypeRSA)
	if err != nil {
		b.Fatalf("Failed to create SSH key: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := ssh.AddSSHKeyToAgent(keyFile)
		if err != nil {
			b.Fatalf("Failed to add SSH key to agent: %v", err)
		}
	}
}
