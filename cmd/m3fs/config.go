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
workDir: "/opt/3fs"
networkType: "RDMA"
nodes:
  - name: node1
    host: "192.168.1.1"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.1"
  - name: node2
    host: "192.168.1.2"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.2"
  - name: node3
    host: "192.168.1.3"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.3"
services:
  fdb:
    nodes: 
      - node1
  clickhouse:
    nodes: 
      - node1
  monitor:
    nodes:
      - node1
  mgmtd:
    nodes: 
      - node1
  meta:
    nodes: 
      - node1
  storage:
    nodes: 
      - node2
      - node3
    diskType: "NVMe"
  client:
    nodes: 
      - node1
    hostMountpoint: /mnt/3fs
images:
  registry: "{{ .registry }}"
  3fs:
    repo: "open3fs/3fs"
    tag: "20250315"
  fdb: 
    repo: "open3fs/foundationdb"
    tag: "7.3.63"
  clickhouse:
    repo: "open3fs/clickhouse"
    tag: "25.1-jammy"
`

func createSampleConfig(ctx *cli.Context) error {
	registry := ""
	if os.Getenv("M3FS_ZONE") == "CN" {
		// TODO: set to cn registry
		registry = ""
	}

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
	err = tmpl.Execute(file, map[string]string{
		"name":     clusterName,
		"registry": registry,
	})
	if err != nil {
		return errors.Annotate(err, "write sample config file")
	}

	return nil
}
