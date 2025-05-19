# MultiGit

MultiGit is a command-line tool that simplifies managing multiple GitHub accounts. It handles SSH key management and automatically configures git settings for seamless context switching between accounts.

[![Go](https://github.com/cpuix/multigit/actions/workflows/test.yml/badge.svg)](https://github.com/cpuix/multigit/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/cpuix/multigit/graph/badge.svg?token=YOUR-TOKEN)](https://codecov.io/gh/cpuix/multigit)
[![Go Report Card](https://goreportcard.com/badge/github.com/cpuix/multigit)](https://goreportcard.com/report/github.com/cpuix/multigit)
[![Test Coverage](https://img.shields.io/badge/coverage-45.7%25-yellow)](https://github.com/cpuix/multigit/actions)

## Features

- 🚀 Create and manage SSH keys for multiple GitHub accounts
- 🔄 Quick switching between different accounts
- 📊 Profile management for account grouping
- 🔒 Secure SSH key management
- ⚡ Automatic SSH config file management
- 🎨 Colorful and user-friendly interface
- 🟡 45.7% test coverage (actively improving)

## 📦 Installation

### macOS (Homebrew)

```bash
# Add custom tap (first time only)
brew tap cpuix/multigit

# Install MultiGit
brew install multigit
```

### Linux (DEB/RPM)

```bash
# DEB (Ubuntu/Debian)
wget https://github.com/cpuix/multigit/releases/latest/download/multigit_linux_amd64.deb
sudo dpkg -i multigit_linux_amd64.deb

# RPM (Fedora/CentOS/RHEL)
wget https://github.com/cpuix/multigit/releases/latest/download/multigit_linux_amd64.rpm
sudo rpm -i multigit_linux_amd64.rpm
```

### Windows

1. [Download the latest release](https://github.com/cpuix/multigit/releases/latest)
2. Run the downloaded `.msi` file
3. Follow the installation wizard

### Using with Docker

```bash
# Run MultiGit
docker run --rm -it -v ~/.ssh:/root/.ssh -v $(pwd):/workspace ghcr.io/cpuix/multigit:latest

# Or add an alias
echo 'alias multigit="docker run --rm -it -v ~/.ssh:/root/.ssh -v $(pwd):/workspace ghcr.io/cpuix/multigit:latest"' >> ~/.bashrc
```

### Go Installation (For Developers)

1. Install Go (1.21 or later)
2. Install MultiGit:

```bash
go install github.com/cpuix/multigit@latest
```

### Manual Installation (Binary)

1. [Download the latest release](https://github.com/cpuix/multigit/releases/latest)
2. Add the binary to your PATH
3. Make it executable:

```bash
chmod +x multigit
sudo mv multigit /usr/local/bin/
```

## Quick Start

### Adding a New Account

```bash
multigit create <account_name> <email@example.com>
```

Example:
```bash
multigit create work-account name.surname@company.com
multigit create personal me@mydomain.com -p "strong-password"
```

### Switching Between Accounts

```bash
multigit use <account_name>
```

### Profile Management

```bash
# Create a new profile
multigit profile create <profile_name>

# List all profiles
multigit profile list

# Use a specific profile
multigit profile use <profile_name>

# Delete a profile
multigit profile delete <profile_name>
```

### Other Commands

```bash
# List all accounts
multigit list

# Show active account
multigit status

# Delete an account
multigit delete <account_name>
```

## Development

### Running Tests

```bash
# Run all tests
make test

# View test coverage
make cover

# Run linter
make lint
```

## License

This project is licensed under the [MIT License](LICENSE).
