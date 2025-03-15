package mgmtd

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenAdminCliConfigSuite(t *testing.T) {
	suiteRun(t, &genAdminCliConfigStepSuite{})
}

type genAdminCliConfigStepSuite struct {
	ttask.StepSuite

	step *genAdminCliConfigStep
}

func (s *genAdminCliConfigStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.Cfg.Name = "test-cluster"
	s.SetupRuntime()
	s.step = &genAdminCliConfigStep{}
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
}

func (s *genAdminCliConfigStepSuite) Test() {
	s.NoError(s.step.Execute(s.Ctx()))

	dataI, ok := s.Runtime.Load(task.RuntimeAdminCliTomlKey)
	s.True(ok)
	s.Equal([]byte(`cluster_id = "test-cluster"

[fdb]
clusterFile = '/opt/3fs/etc/fdb.cluster'`), dataI.([]byte))
}

func TestInitClusterStepSuite(t *testing.T) {
	suiteRun(t, &initClusterStepSuite{})
}

type initClusterStepSuite struct {
	ttask.StepSuite

	step      *initClusterStep
	configDir string
}

func (s *initClusterStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &initClusterStep{}
	s.Cfg.Services.Mgmtd.WorkDir = "/var/mgmtd"
	s.configDir = "/var/mgmtd/config.d"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, "xxxx")
}

func (s *initClusterStepSuite) TestInitCluster() {
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "3fs")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:      img,
		Name:       &s.Cfg.Services.Mgmtd.ContainerName,
		Entrypoint: common.Pointer("''"),
		Rm:         common.Pointer(true),
		Command: []string{
			"/opt/3fs/bin/admin_cli",
			"-cfg", "/opt/3fs/etc/admin_cli.toml",
			"'init-cluster --mgmtd /opt/3fs/etc/mgmtd_main.toml 1 1048576 16'",
		},
		HostNetwork: true,
		Volumes: []*external.VolumeArgs{
			{
				Source: s.configDir,
				Target: "/opt/3fs/etc",
			},
		},
	}).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func TestInitUserAndChainSuite(t *testing.T) {
	suiteRun(t, &initUserAndChainStepSuite{})
}

type initUserAndChainStepSuite struct {
	ttask.StepSuite

	step *initUserAndChainStep
}

func (s *initUserAndChainStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &initUserAndChainStep{}
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeMgmtdServerAddressesKey, `["RDMA://10.16.28.58:8000"]`)
}

func (s *initUserAndChainStepSuite) Test() {
	containerName := s.Runtime.Services.Mgmtd.ContainerName
	s.MockDocker.On("Exec", containerName, "/opt/3fs/bin/admin_cli", []string{
		"-cfg", "/opt/3fs/etc/admin_cli.toml",
		"--config.mgmtd_client.mgmtd_server_addresses", `'["RDMA://10.16.28.58:8000"]'`,
		`"user-add --root --admin 0 root"`,
	}).Return(`Uid                0
Name               root
Token              AAA8WCoB8QAt8bFw2wBupzjA(Expired at N/A)
IsRootUser         true
IsAdmin            true
Gid                0
SupplementaryGids`, nil)
	s.MockDocker.On("Exec", containerName, "bash", []string{
		"-c",
		`"echo AAA8WCoB8QAt8bFw2wBupzjA > /opt/3fs/etc/token.txt"`,
	}).Return("", nil)

	s.MockDocker.On("Exec", containerName, "python3", []string{
		"/opt/3fs/data_placement/src/model/data_placement.py",
		"-ql", "-relax", "-type", "CR",
		"--num_nodes", strconv.Itoa(len(s.Runtime.Services.Storage.Nodes)),
		"--replication_factor", strconv.Itoa(s.Runtime.Services.Storage.ReplicationFactor),
		"--min_targets_per_disk", strconv.Itoa(s.Runtime.Services.Storage.MinTargetNumPerDisk),
	}).Return(`2025-03-15 08:34:31.409 | INFO     | __main__:check_solution:332`+
		`- total_traffic=64.0 max_total_traffic=64
		2025-03-15 08:34:31.640 | SUCCESS  | __main__:run:148 - `+
		`saved solution to: output/DataPlacementModel-v_2-b_32-r_32-k_2-λ_32-lb_1-ub_0`, nil)
	s.MockDocker.On("Exec", containerName, "python3", []string{
		"/opt/3fs/data_placement/src/setup/gen_chain_table.py",
		"--chain_table_type", "CR",
		"--node_id_begin", "10001",
		"--node_id_end", strconv.Itoa(10000 + len(s.Runtime.Services.Storage.Nodes)),
		"--num_disks_per_node", strconv.Itoa(s.Runtime.Services.Storage.DiskNumPerNode),
		"--num_targets_per_disk", strconv.Itoa(s.Runtime.Services.Storage.TargetNumPerDisk),
		"--target_id_prefix", strconv.Itoa(s.Runtime.Services.Storage.TargetIDPrefix),
		"--chain_id_prefix", strconv.Itoa(s.Runtime.Services.Storage.ChainIDPrefix),
		"--incidence_matrix_path", "output/DataPlacementModel-v_2-b_32-r_32-k_2-λ_32-lb_1-ub_0/incidence_matrix.pickle",
	}).Return("", nil)
	s.MockDocker.On("Exec", containerName, "bash", []string{
		"-c",
		`"/opt/3fs/bin/admin_cli --cfg /opt/3fs/etc/admin_cli.toml ` +
			`--config.mgmtd_client.mgmtd_server_addresses '[\"RDMA://10.16.28.58:8000\"]' ` +
			`--config.user_info.token AAA8WCoB8QAt8bFw2wBupzjA < output/create_target_cmd.txt"`,
	}).Return("", nil)
	s.MockDocker.On("Exec", containerName, "/opt/3fs/bin/admin_cli", []string{
		"--cfg", "/opt/3fs/etc/admin_cli.toml",
		"--config.mgmtd_client.mgmtd_server_addresses", `'["RDMA://10.16.28.58:8000"]'`,
		"--config.user_info.token", "AAA8WCoB8QAt8bFw2wBupzjA",
		`"upload-chains output/generated_chains.csv"`,
	}).Return("", nil)
	s.MockDocker.On("Exec", containerName, "/opt/3fs/bin/admin_cli", []string{
		"--cfg", "/opt/3fs/etc/admin_cli.toml",
		"--config.mgmtd_client.mgmtd_server_addresses", `'["RDMA://10.16.28.58:8000"]'`,
		"--config.user_info.token", "AAA8WCoB8QAt8bFw2wBupzjA",
		`"upload-chain-table --desc stage 1 output/generated_chain_table.csv"`,
	}).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))
}
