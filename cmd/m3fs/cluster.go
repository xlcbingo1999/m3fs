// Copyright 2025 Open3FS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	fsclient "github.com/open3fs/m3fs/pkg/3fs_client"
	"github.com/open3fs/m3fs/pkg/clickhouse"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/fdb"
	"github.com/open3fs/m3fs/pkg/meta"
	"github.com/open3fs/m3fs/pkg/mgmtd"
	"github.com/open3fs/m3fs/pkg/monitor"
	"github.com/open3fs/m3fs/pkg/network"
	"github.com/open3fs/m3fs/pkg/storage"
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
		{
			Name:   "prepare",
			Usage:  "prepare to deploy a 3fs cluster",
			Action: prepareCluster,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "Path to the cluster configuration file",
					Destination: &configFilePath,
					Required:    true,
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

	runner := task.NewRunner(cfg,
		new(fdb.CreateFdbClusterTask),
		new(clickhouse.CreateClickhouseClusterTask),
		new(monitor.CreateMonitorTask),
		new(mgmtd.CreateMgmtdServiceTask),
		new(meta.CreateMetaServiceTask),
		new(storage.CreateStorageServiceTask),
		new(mgmtd.InitUserAndChainTask),
		new(fsclient.Create3FSClientServiceTask),
	)
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

	runner := task.NewRunner(cfg,
		new(fsclient.Delete3FSClientServiceTask),
		new(storage.DeleteStorageServiceTask),
		new(meta.DeleteMetaServiceTask),
		new(mgmtd.DeleteMgmtdServiceTask),
		new(monitor.DeleteMonitorTask),
		new(clickhouse.DeleteClickhouseClusterTask),
		new(fdb.DeleteFdbClusterTask),
	)
	runner.Init()
	if err = runner.Run(ctx.Context); err != nil {
		return errors.Annotate(err, "delete cluster")
	}

	return nil
}

func prepareCluster(ctx *cli.Context) error {
	cfg, err := loadClusterConfig()
	if err != nil {
		return errors.Trace(err)
	}
	runner := task.NewRunner(cfg,
		new(network.PrepareNetworkTask),
	)
	runner.Init()
	if err = runner.Run(ctx.Context); err != nil {
		return errors.Annotate(err, "prepare cluster")
	}

	return nil
}
