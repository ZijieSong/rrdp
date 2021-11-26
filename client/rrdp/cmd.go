package main

import (
	"../../client"
	"../../common"
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
		LocalPorts: &cli.StringSlice{},
	}

	app := cli.NewApp()
	app.Name = "net proxy tools"
	app.Usage = ""
	app.Version = version
	app.Flags = client.AppFlags(options)
	app.Commands = []*cli.Command{
		newLocalToRemoteCommand(options),
	}

	var cleaner common.Cleaner = &client.CliCleaner{}
	common.SetupCloseHandler(&cleaner)
	defer cleaner.Clean()

	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msgf("End with error: %s", err.Error())
	}
}

func newLocalToRemoteCommand(options *client.CliOptions) *cli.Command {
	return &cli.Command{
		Name:  "lToR",
		Usage: "local to remote proxy",
		Action: func(context *cli.Context) error {
			return client.Main(options)
		},
	}
}
