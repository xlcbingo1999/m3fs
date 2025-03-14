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
workDir: "/root/3fs"
networktype: "RDMA"
nodes:
  - name: meta
    ip: "192.168.1.1"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.1"
  - name: storage1
    ip: "192.168.1.2"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.2"
  - name: storage2
    ip: "192.168.1.3"
    username: "root"
    password: "password"
    rdmaAddresses:
      - "10.0.0.3"
services:
  fdb:
    containerName: 3fs-fdb
    nodes: 
      - meta
  clickhouse:
    containerName: 3fs-clickhouse
    nodes: 
      - meta
    db: 3fs
    user: default
    password: password
    tcpPort: 9000
  monitor:
    containerName: 3fs-monitor
    nodes:
      - meta
  mgmtd:
    containerName: 3fs-mgmtd
    nodes: 
      - meta
  meta:
    containerName: 3fs-meta
    nodes: 
      - meta
  storage:
    containerName: 3fs-storage
    nodes: 
      - storage1
      - storage2
    disktype: "NVMe"
  client:
    containerName: 3fs-fuseclient
    nodes: 
      - meta
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
