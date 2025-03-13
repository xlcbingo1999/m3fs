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
}

// Mgmtd is the 3fs mgmtd service config definition
type Mgmtd struct {
	ContainerName  string `yaml:"containerName"`
	Nodes          []string
	WorkDir        string `yaml:"workDir"`
	RDMAListenPort int    `yaml:"rdmaListenPort"`
	TCPListenPort  int    `yaml:"tcpListenPort"`
}

// Meta is the 3fs meta service config definition
type Meta struct {
	ContainerName string `yaml:"containerName"`
	Nodes         []string
}

// Storage is the 3fs storage config definition
type Storage struct {
	ContainerName string `yaml:"containerName"`
	Nodes         []string
	DiskType      DiskType
}

// Client is the 3fs client config definition
type Client struct {
	ContainerName string `yaml:"containerName"`
	Nodes         []string
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

// Registry is the service contaner image registry config definition
type Registry struct {
	CustomRegistry string
}

// Config is the 3fs cluster config definition
type Config struct {
	Name             string
	WorkDir          string `yaml:"workDir"`
	NetworkType      NetworkType
	Nodes            []Node
	Services         Services `yaml:"services"`
	Registry         Registry
	CmdMaxExitTimout *time.Duration `yaml:",omitempty"`
}

// SetValidate validates the config and set default values if some fields are missing
func (c *Config) SetValidate(workDir string) error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	c.WorkDir = workDir
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
	if err := c.validServiceNodes("fdb", c.Services.Fdb.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}
	if c.Services.Fdb.Port == 0 {
		c.Services.Fdb.Port = 4500
	}
	if c.Services.Fdb.WorkDir == "" {
		c.Services.Fdb.WorkDir = path.Join(workDir, "fdb")
	}

	if err := c.validServiceNodes("clickhouse", c.Services.Clickhouse.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}
	if c.Services.Clickhouse.WorkDir == "" {
		c.Services.Clickhouse.WorkDir = path.Join(workDir, "clickhouse")
	}

	if err := c.validServiceNodes("monitor", c.Services.Monitor.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}
	if c.Services.Monitor.WorkDir == "" {
		c.Services.Monitor.WorkDir = path.Join(workDir, "monitor")
	}

	if err := c.validServiceNodes("mgmtd", c.Services.Mgmtd.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}

	if err := c.validServiceNodes("meta", c.Services.Meta.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}

	if err := c.validServiceNodes("storag", c.Services.Storage.Nodes, nodeSet, true); err != nil {
		return errors.Trace(err)
	}
	if !diskTypes.Contains(c.Services.Storage.DiskType) {
		return errors.Errorf("invalid disk type of storage service: %s", c.Services.Storage.DiskType)
	}

	if err := c.validServiceNodes("client", c.Services.Client.Nodes, nodeSet, false); err != nil {
		return errors.Trace(err)
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
			},
			Monitor: Monitor{
				ContainerName: "3fs-monitor",
			},
			Mgmtd: Mgmtd{
				ContainerName: "3fs-mgmtd",
			},
			Meta: Meta{
				ContainerName: "3fs-meta",
			},
			Storage: Storage{
				ContainerName: "3fs-storage",
				DiskType:      DiskTypeNvme,
			},
			Client: Client{
				ContainerName: "3fs-client",
			},
		},
	}
}
