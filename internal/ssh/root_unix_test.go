//go:build !windows

package ssh_test

import "os"

func isRootUser() bool {
	return os.Geteuid() == 0
}
