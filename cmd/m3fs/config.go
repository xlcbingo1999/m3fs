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
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/open3fs/m3fs/pkg/errors"
)

var (
	clusterName      string
	sampleConfigPath string
)

var configCmd = &cli.Command{
	Name:    "config",
	Aliases: []string{"cfg"},
	Usage:   "Manage 3fs config",
	Subcommands: []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a sample 3fs config",
			Action: createSampleConfig,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "name",
					Aliases:     []string{"n"},
					Usage:       "3FS cluster name",
					Value:       "3fs",
					Destination: &clusterName,
				},
				&cli.StringFlag{
					Name:        "file",
					Aliases:     []string{"f"},
					Usage:       "Specify a configuration file path",
					Destination: &sampleConfigPath,
					Value:       "cluster_sample.yml",
				},
			},
		},
	},
}

var sampleConfigTemplate = `name: "{{.name}}"
networktype: "RDMA"
nodes:
  - name: meta
    host: "192.168.1.1"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.1"
  - name: storage1
    host: "192.168.1.2"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.2"
  - name: storage2
    host: "192.168.1.3"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.3"
services:
  fdb:
    nodes: 
      - meta
  clickhouse:
    nodes: 
      - meta
  monitor:
    nodes:
      - meta
  mgmtd:
    nodes: 
      - meta
  meta:
    nodes: 
      - meta
  storage:
    nodes: 
      - storage1
      - storage2
    diskType: "NVMe"
  client:
    nodes: 
      - meta
    hostMountpoint: /mnt/3fs
registry:
  customRegistry: ""
`

func createSampleConfig(ctx *cli.Context) error {
	tmpl, err := template.New("sampleConfig").Parse(sampleConfigTemplate)
	if err != nil {
		return errors.Annotate(err, "parse sample config template")
	}
	if clusterName == "" {
		return errors.New("cluster name is required")
	}
	if sampleConfigPath == "" {
		sampleConfigPath = "cluster_sample.yml"
	}

	file, err := os.OpenFile(sampleConfigPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return errors.Annotate(err, "create sample config file")
	}
	err = tmpl.Execute(file, map[string]string{"name": clusterName})
	if err != nil {
		return errors.Annotate(err, "write sample config file")
	}

	return nil
}
