package client

import (
	"os"
	"os/exec"
)

type CliCleaner struct {
}

func (c *CliCleaner) Clean() error {
	//clean nat rules
	cmd := exec.Command("pfctl", "-ef", "/etc/pf.conf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
