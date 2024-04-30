package multigit

import "errors"

// add
// delete
// copy
// viewaccounts (list)

type MultiGitExecutor interface {
	Exec() error
	Add() error
	Create() error
	Delete() error
	Copy() error
	Help() error
	UseAccount() error
	ViewAccounts()
}

type MultiGit struct {
	Cmd  string
	Args []string
}

// UseAccount implements MultiGitExecutor.
func (m *MultiGit) UseAccount() error {
	panic("unimplemented")
}

// Create implements MultiGitExecutor.
func (m *MultiGit) Create() error {
	panic("unimplemented")
}

// Help implements MultiGitExecutor.
func (m *MultiGit) Help() error {
	panic("unimplemented")
}

var (
	Add          = "add"
	Delete       = "delete"
	Copy         = "copy"
	ViewAccounts = "viewaccounts"
)

// Exec implements MultiGitExecutor.
func (m *MultiGit) Exec() error {
	switch m.Cmd {
	case Add:
		return m.Add()
	case Delete:
		return m.Delete()
	case Copy:
		return m.Copy()
	case ViewAccounts:
		m.ViewAccounts()
	default:
		return errors.New("Unknown execute method!")
		// panic("Unknown execute method!")
	}
	return nil
}

// Add implements MultiGitExecutor.
func (m *MultiGit) Add() error {
	panic("unimplemented")
}

// Copy implements MultiGitExecutor.
func (m *MultiGit) Copy() error {
	panic("unimplemented")
}

// Delete implements MultiGitExecutor.
func (m *MultiGit) Delete() error {
	panic("unimplemented")
}

// ViewAccounts implements MultiGitExecutor.
func (m *MultiGit) ViewAccounts() {
	panic("unimplemented")
}

// MultiGit implements MultiGitExecutor
var _ MultiGitExecutor = &MultiGit{}
