package config

import "os"

// SSHConfigPath const SSHConfigPath = "~/.ssh/config"
const SSHConfigPath = "../../config"

func AppendConfigEntry(entry string) error {
	file, err := os.OpenFile(SSHConfigPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(entry); err != nil {
		return err
	}
	return nil
}
