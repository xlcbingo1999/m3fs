package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/fdb"
	"github.com/open3fs/m3fs/pkg/task"
)

var (
	configFilePath string
	workDir        string
)

var clusterCmd = &cli.Command{
	Name:    "cluster",
	Aliases: []string{"c"},
	Usage:   "Manage 3fs cluster",
	Subcommands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new 3fs cluster",
			Action: createCluster,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "Path to the cluster configuration file",
					Destination: &configFilePath,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "workdir",
					Aliases:     []string{"w"},
					Usage:       "Path to the working directory(default is current directory)",
					Destination: &workDir,
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "delete a 3fs cluster",
			Action: deleteCluster,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "Path to the cluster configuration file",
					Destination: &configFilePath,
					Required:    true,
				},
				&cli.StringFlag{
					Name:        "workdir",
					Aliases:     []string{"w"},
					Usage:       "Path to the working directory(default is current directory)",
					Destination: &workDir,
				},
			},
		},
	},
}

func loadClusterConfig() (*config.Config, error) {
	cfg := config.NewConfigWithDefaults()
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, errors.Annotate(err, "open config file")
	}
	if err = yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, errors.Annotate(err, "load cluster config")
	}
	if err = cfg.SetValidate(workDir); err != nil {
		return nil, errors.Annotate(err, "validate cluster config")
	}
	logrus.Debugf("Cluster config: %+v", cfg)

	return cfg, nil
}

func createCluster(ctx *cli.Context) error {
	cfg, err := loadClusterConfig()
	if err != nil {
		return errors.Trace(err)
	}

	runner := task.NewRunner(cfg, new(fdb.CreateFdbClusterTask))
	runner.Init()
	if err = runner.Run(ctx.Context); err != nil {
		return errors.Annotate(err, "create cluster")
	}

	return nil
}

func deleteCluster(ctx *cli.Context) error {
	cfg, err := loadClusterConfig()
	if err != nil {
		return errors.Trace(err)
	}

	runner := task.NewRunner(cfg, new(fdb.DeleteFdbClusterTask))
	runner.Init()
	if err = runner.Run(ctx.Context); err != nil {
		return errors.Annotate(err, "delete cluster")
	}

	return nil
}
