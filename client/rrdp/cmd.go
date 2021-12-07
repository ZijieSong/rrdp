package main

import (
	"github.com/ZijieSong/rrdp/client"
	"github.com/ZijieSong/rrdp/common"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	"os"
)

var version = "0.0.1"

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	options := &client.CliOptions{
		LocalPorts:   &cli.StringSlice{},
		ConfigPath:   os.Getenv("HOME"),
		ExposedPorts: &cli.StringSlice{},
	}

	app := cli.NewApp()
	app.Name = "net proxy tools"
	app.Usage = ""
	app.Version = version
	app.Flags = client.AppFlags(options)
	app.Commands = []*cli.Command{
		newLocalToRemoteCommand(options),
		newRemoteToLocalCommand(options),
	}

	var cleaner common.Cleaner = &client.CliCleaner{}
	app.ExitErrHandler = func(_ *cli.Context, err error) {
		log.Error().Msgf("End with error %s", err.Error())
		err = cleaner.Clean()
		if err != nil {
			log.Error().Msgf("cleaner occurs error, %s", err.Error())
		}
		os.Exit(1)
	}
	common.SetupCloseHandler(&cleaner)

	_ = app.Run(os.Args)
}

func newLocalToRemoteCommand(options *client.CliOptions) *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "local to remote proxy",
		Flags: client.LocalToRemote(options),
		Action: func(context *cli.Context) error {
			return client.Proxy(options)
		},
	}
}

func newRemoteToLocalCommand(options *client.CliOptions) *cli.Command {
	return &cli.Command{
		Name:  "expose",
		Usage: "remote to local proxy",
		Flags: client.RemoteToLocal(options),
		Action: func(context *cli.Context) error {
			return client.Expose(options)
		},
	}
}
