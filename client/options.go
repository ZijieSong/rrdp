package client

import (
	"github.com/urfave/cli"
)

type CliOptions struct {
	RemoteServer string

	LocalPorts *cli.StringSlice
	ConfigPath string

	ExposedPorts *cli.StringSlice
}
