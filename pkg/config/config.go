package config

import (
	"path"
	"time"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/utils"
)

// NetworkType is the type of network definition
type NetworkType string

// defines network types
const (
	NetworkTypeRDMA NetworkType = "RDMA"
	NetworkTypeRXE  NetworkType = "RXE"
)

var networkTypes = utils.NewSet(NetworkTypeRDMA, NetworkTypeRXE)

// DiskType is the type of disk definition
type DiskType string

// defines disk types
const (
	DiskTypeDirectory DiskType = "directory"
	DiskTypeNvme      DiskType = "NVMe"
)

var diskTypes = utils.NewSet(DiskTypeDirectory, DiskTypeNvme)

// Node is the node config definition
type Node struct {
	Name          string
	Host          string
	Port          int
	Username      string
	Password      *string  `yaml:",omitempty"`
	RDMAAddresses []string `yaml:"rdmaAddresses,omitempty"`
}

// Fdb is the fdb config definition
type Fdb struct {
	ContainerName      string `yaml:"containerName"`
	Nodes              []string
	Port               int
	WorkDir            string `yaml:"workDir"`
	WaitClusterTimeout time.Duration
}

// Clickhouse is the click house config definition
type Clickhouse struct {
	ContainerName string `yaml:"containerName"`
	Nodes         []string
	WorkDir       string `yaml:"workDir"`
	Db            string `yaml:"db"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	TCPPort       int    `yaml:"tcpPort"`
}

// Monitor is the monitor config definition
type Monitor struct {
	ContainerName string `yaml:"containerName"`
	Nodes         []string
	WorkDir       string `yaml:"workDir"`
	Port          int    `yaml:"port"`
}

// Mgmtd is the 3fs mgmtd service config definition
type Mgmtd struct {
	ContainerName  string `yaml:"containerName"`
	Nodes          []string
	WorkDir        string `yaml:"workDir,omitempty"`
	RDMAListenPort int    `yaml:"rdmaListenPort,omitempty"`
	TCPListenPort  int    `yaml:"tcpListenPort,omitempty"`
}

// Meta is the 3fs meta service config definition
type Meta struct {
	ContainerName  string `yaml:"containerName"`
	Nodes          []string
	WorkDir        string `yaml:"workDir,omitempty"`
	RDMAListenPort int    `yaml:"rdmaListenPort,omitempty"`
	TCPListenPort  int    `yaml:"tcpListenPort,omitempty"`
}

// Storage is the 3fs storage config definition
type Storage struct {
	ContainerName       string `yaml:"containerName"`
	Nodes               []string
	WorkDir             string   `yaml:"workDir,omitempty"`
	DiskType            DiskType `yaml:"diskType,omitempty"`
	DiskNumPerNode      int      `yaml:"diskNumPerNode,omitempty"`
	RDMAListenPort      int      `yaml:"rdmaListenPort,omitempty"`
	TCPListenPort       int      `yaml:"tcpListenPort,omitempty"`
	ReplicationFactor   int      `yaml:"replicationFactor,omitempty"`
	MinTargetNumPerDisk int      `yaml:"minTargetNumPerDisk,omitempty"`
	TargetNumPerDisk    int      `yaml:"targetNumPerDisk,omitempty"`
	TargetIDPrefix      int      `yaml:"targetIDPrefix,omitempty"`
	ChainIDPrefix       int      `yaml:"chainIDPrefix,omitempty"`
}

// Client is the 3fs client config definition
type Client struct {
	ContainerName  string `yaml:"containerName"`
	Nodes          []string
	HostMountpoint string `yaml:"hostMountpoint"`
}

// Services is the services config definition
type Services struct {
	Fdb        Fdb
	Clickhouse Clickhouse
	Monitor    Monitor
	Mgmtd      Mgmtd
	Meta       Meta
	Storage    Storage
	Client     Client
}

// Registry is the service container image registry config definition
type Registry struct {
	CustomRegistry string `yaml:"customRegistry"`
}

// Config is the 3fs cluster config definition
type Config struct {
	Name              string
	WorkDir           string `yaml:"workDir"`
	NetworkType       NetworkType
	Nodes             []Node
	Services          Services `yaml:"services"`
	Registry          Registry
	CmdMaxExitTimeout *time.Duration `yaml:",omitempty"`
}

// SetValidate validates the config and set default values if some fields are missing
func (c *Config) SetValidate(workDir string) error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if workDir != "" {
		c.WorkDir = workDir
	}
	if !networkTypes.Contains(c.NetworkType) {
		return errors.Errorf("invalid network type: %s", c.NetworkType)
	}
	if len(c.Nodes) == 0 {
		return errors.New("nodes is required")
	}
	nodeSet := utils.NewSet[string]()
	nodeHostSet := utils.NewSet[string]()
	for i, node := range c.Nodes {
		if node.Name == "" {
			return errors.Errorf("nodes[%d].name is required", i)
		}
		if !nodeSet.AddIfNotExists(node.Name) {
			return errors.Errorf("duplicate node name: %s", node.Name)
		}
		if node.Host == "" {
			return errors.Errorf("nodes[%d].host is required", i)
		}
		if !nodeHostSet.AddIfNotExists(node.Host) {
			return errors.Errorf("duplicate node host: %s", node.Host)
		}
		if node.Port == 0 {
			c.Nodes[i].Port = 22
		}
	}

	validSettings := []struct {
		name    string
		nodes   []string
		require bool
	}{
		{
			"fdb",
			c.Services.Fdb.Nodes,
			true,
		},
		{
			"clickhouse",
			c.Services.Clickhouse.Nodes,
			true,
		},
		{
			"monitor",
			c.Services.Monitor.Nodes,
			true,
		},
		{
			"mgmtd",
			c.Services.Mgmtd.Nodes,
			true,
		},
		{
			"meta",
			c.Services.Mgmtd.Nodes,
			true,
		},
		{
			"storage",
			c.Services.Storage.Nodes,
			true,
		},
		{
			"client",
			c.Services.Client.Nodes,
			false,
		},
	}

	for _, s := range validSettings {
		if err := c.validServiceNodes(s.name, s.nodes, nodeSet, s.require); err != nil {
			return errors.Trace(err)
		}
	}

	if c.Services.Fdb.Port == 0 {
		c.Services.Fdb.Port = 4500
	}
	if c.Services.Fdb.WorkDir == "" {
		c.Services.Fdb.WorkDir = path.Join(c.WorkDir, "fdb")
	}

	if c.Services.Clickhouse.WorkDir == "" {
		c.Services.Clickhouse.WorkDir = path.Join(c.WorkDir, "clickhouse")
	}

	if c.Services.Monitor.WorkDir == "" {
		c.Services.Monitor.WorkDir = path.Join(c.WorkDir, "monitor")
	}

	if c.Services.Mgmtd.WorkDir == "" {
		c.Services.Mgmtd.WorkDir = path.Join(c.WorkDir, "mgmtd")
	}
	if c.Services.Mgmtd.RDMAListenPort == 0 {
		c.Services.Mgmtd.RDMAListenPort = 8000
	}
	if c.Services.Mgmtd.TCPListenPort == 0 {
		c.Services.Mgmtd.TCPListenPort = 9000
	}

	if c.Services.Meta.WorkDir == "" {
		c.Services.Meta.WorkDir = path.Join(c.WorkDir, "meta")
	}
	if c.Services.Meta.RDMAListenPort == 0 {
		c.Services.Meta.RDMAListenPort = 8001
	}
	if c.Services.Meta.TCPListenPort == 0 {
		c.Services.Meta.TCPListenPort = 9001
	}

	if !diskTypes.Contains(c.Services.Storage.DiskType) {
		return errors.Errorf("invalid disk type of storage service: %s", c.Services.Storage.DiskType)
	}
	if c.Services.Storage.WorkDir == "" {
		c.Services.Storage.WorkDir = path.Join(c.WorkDir, "storage")
	}
	if c.Services.Storage.RDMAListenPort == 0 {
		c.Services.Storage.RDMAListenPort = 8002
	}
	if c.Services.Storage.TCPListenPort == 0 {
		c.Services.Storage.TCPListenPort = 9002
	}
	if c.Services.Storage.DiskNumPerNode == 0 {
		c.Services.Storage.DiskNumPerNode = 1
	}

	if c.Services.Client.HostMountpoint == "" {
		return errors.New("services.client.hostMountpoint is required")
	}

	return nil
}

func (c *Config) validServiceNodes(
	service string, nodes []string, nodeSet *utils.Set[string], required bool) error {

	if required && len(nodes) == 0 {
		return errors.Errorf("nodes of %s service is required", service)
	}
	serviceNodeSet := utils.NewSet[string]()
	for _, node := range nodes {
		if !nodeSet.Contains(node) {
			return errors.Errorf("node %s of  %s service not exists in node list", node, service)
		}
		if !serviceNodeSet.AddIfNotExists(node) {
			return errors.Errorf("duplicate node %s in %s service", node, service)
		}
	}
	return nil
}

// NewConfigWithDefaults creates a new config with default values
func NewConfigWithDefaults() *Config {
	return &Config{
		Name:        "3fs",
		NetworkType: NetworkTypeRDMA,
		Services: Services{
			Fdb: Fdb{
				ContainerName:      "3fs-fdb",
				Port:               4500,
				WaitClusterTimeout: 120 * time.Second,
			},
			Clickhouse: Clickhouse{
				ContainerName: "3fs-clickhouse",
				TCPPort:       8999,
			},
			Monitor: Monitor{
				ContainerName: "3fs-monitor",
				Port:          10000,
			},
			Mgmtd: Mgmtd{
				ContainerName:  "3fs-mgmtd",
				RDMAListenPort: 8000,
				TCPListenPort:  9000,
			},
			Meta: Meta{
				ContainerName: "3fs-meta",
			},
			Storage: Storage{
				ContainerName:       "3fs-storage",
				DiskType:            DiskTypeNvme,
				ReplicationFactor:   2,
				DiskNumPerNode:      1,
				MinTargetNumPerDisk: 32,
				TargetNumPerDisk:    32,
				TargetIDPrefix:      1,
				ChainIDPrefix:       9,
			},
			Client: Client{
				ContainerName:  "3fs-client",
				HostMountpoint: "/mnt/3fs",
			},
		},
	}
}
