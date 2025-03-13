package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/open3fs/m3fs/pkg/errors"
)

var debug bool

func initLogger() {
	if debug {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	}
}

func main() {
	app := &cli.App{
		Name:  "m3fs",
		Usage: "3FS Deploy Tool",
		Before: func(ctx *cli.Context) error {
			initLogger()
			return nil
		},
		Commands: []*cli.Command{
			clusterCmd,
			configCmd,
			osCmd,
			tmplCmd,
		},
		Action: func(ctx *cli.Context) error {
			return cli.ShowAppHelp(ctx)
		},
		ExitErrHandler: func(cCtx *cli.Context, err error) {
			if err != nil {
				logrus.Debugf("Command failed stacktrace: %s", errors.StackTrace(err))
			}
			cli.HandleExitCoder(err)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable debug mode",
				Destination: &debug,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
