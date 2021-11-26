package client

import (
	"github.com/urfave/cli"
)

// AppFlags return app flags
func AppFlags(options *CliOptions) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "remote,r",
			Usage:       "Specify remote proxy server",
			Destination: &options.RemoteServer,
		},
		&cli.StringSliceFlag{
			Name:        "localPorts,p",
			Usage:       "Specify local port to be proxy",
			Destination: options.LocalPorts,
		},
	}
}
