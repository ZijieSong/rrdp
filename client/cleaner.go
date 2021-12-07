package client

import (
	"github.com/ZijieSong/rrdp/common"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
)

type CliCleaner struct {
}

func (c *CliCleaner) Clean() error {
	//lock
	flock, err := common.NewFileLock("/tmp/rrdp.lock", true)
	if err != nil {
		return err
	}
	err = flock.Lock()
	if err != nil {
		return err
	}
	log.Info().Msgf("file lock successfully, current pid is %d", os.Getpid())
	defer flock.Unlock()

	//clean nat rules
	cmd := exec.Command("sh", "-c", "pfctl -ef /etc/pf.conf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info().Msgf("clean nat rules by %s", cmd.String())
	if err = cmd.Run(); err != nil {
		return err
	}

	return nil
}
