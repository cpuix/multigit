//go:build !windows

package multigit_test

import "os"

func isRootUser() bool {
	return os.Geteuid() == 0
}
