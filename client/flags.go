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
	}
}

func LocalToRemote(options *CliOptions) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "localPorts,p",
			Usage:       "Specify local port to be proxy",
			Destination: options.LocalPorts,
		},
	}
}

func RemoteToLocal(options *CliOptions) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "exposedPorts,p",
			Usage:       "Specify exposed port, eg: 9090:8080, means remote port is 8080, local port is 9090",
			Destination: options.ExposedPorts,
		},
	}
}
