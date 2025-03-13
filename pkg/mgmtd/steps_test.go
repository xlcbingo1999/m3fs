package mgmtd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/task"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenNodeIDStepSuite(t *testing.T) {
	suiteRun(t, &genNodeIDStepSuite{})
}

type genNodeIDStepSuite struct {
	ttask.StepSuite

	step *genNodeIDStep
}

func (s *genNodeIDStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &genNodeIDStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
		{
			Name: "node2",
			Host: "1.1.1.2",
		},
	}
	s.Cfg.Services.Mgmtd.Nodes = []string{"node1", "node2"}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
}

func (s *genNodeIDStepSuite) TestGenNodeID() {
	s.NoError(s.step.Execute(s.Ctx()))

	idI, ok := s.Runtime.Load(getNodeID(s.Cfg.Nodes[0].Name))
	s.True(ok)
	s.Equal(1, idI.(int))

	idI2, ok := s.Runtime.Load(getNodeID(s.Cfg.Nodes[1].Name))
	s.True(ok)
	s.Equal(2, idI2.(int))
}

func TestPrepareMgmtdConfigStepSuite(t *testing.T) {
	suiteRun(t, &prepareMgmtdConfigStepSuite{})
}

type prepareMgmtdConfigStepSuite struct {
	ttask.StepSuite

	step       *prepareMgmtdConfigStep
	node       config.Node
	fdbContent string
}

func (s *prepareMgmtdConfigStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &prepareMgmtdConfigStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
	}
	s.Cfg.Name = "test-cluster"
	s.node = s.Cfg.Nodes[0]
	s.Cfg.Services.Mgmtd.Nodes = []string{"node1"}
	s.Cfg.Services.Mgmtd.TCPListenPort = 9000
	s.Cfg.Services.Mgmtd.RDMAListenPort = 8000
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
	s.Runtime.Store(getNodeID(s.Cfg.Nodes[0].Name), 1)
	s.fdbContent = "xxxx,xxxxx,xxxx"
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, s.fdbContent)
}

func (s *prepareMgmtdConfigStepSuite) mockGenConfig(path, tmpContent string) {
	var data any = mock.AnythingOfType("[]uint8")
	if tmpContent != "" {
		data = []byte(tmpContent)
	}
	s.MockedLocal.On("WriteFile", path, data, os.FileMode(0644)).Return(nil)
}

func (s *prepareMgmtdConfigStepSuite) getGeneratedConfigContent() (string, string, string) {
	mainApp := `allow_empty_node_id = true
node_id = 1`
	mainLauncher := `allow_dev_version = true
cluster_id = 'test-cluster'
use_memkv = false

[fdb]
casual_read_risky = false
clusterFile = '/opt/3fs/etc/fdb.cluster'
default_backoff = 0
enableMultipleClient = false
externalClientDir = ''
externalClientPath = ''
multipleClientThreadNum = 4
readonly = false
trace_file = ''
trace_format = 'json'

[ib_devices]
allow_no_usable_devices = false
allow_unknown_zone = true
default_network_zone = 'UNKNOWN'
default_pkey_index = 0
default_roce_pkey_index = 0
default_traffic_class = 0
device_filter = []
fork_safe = true
prefer_ibdevice = true
skip_inactive_ports = true
skip_unusable_device = true
subnets = []

[kv_engine]
use_memkv = false

[kv_engine.fdb]
casual_read_risky = false
clusterFile = ''
default_backoff = 0
enableMultipleClient = false
externalClientDir = ''
externalClientPath = ''
multipleClientThreadNum = 4
readonly = false
trace_file = ''
trace_format = 'json'`
	mainContent := `[[common.log.categories]]
categories = [ '.' ]
handlers = [ 'normal', 'err', 'fatal' ]
inherit = true
level = 'INFO'
propagate = 'NONE'

[[common.log.handlers]]
async = true
file_path = '/var/log/3fs/mgmtd_main.log'
max_file_size = '100MB'
max_files = 10
name = 'normal'
rotate = true
rotate_on_open = false
start_level = 'NONE'
stream_type = 'STDERR'
writer_type = 'FILE'

[[common.log.handlers]]
async = false
file_path = '/var/log/3fs/mgmtd_main-err.log'
max_file_size = '100MB'
max_files = 10
name = 'err'
rotate = true
rotate_on_open = false
start_level = 'ERR'
stream_type = 'STDERR'
writer_type = 'FILE'

[[common.log.handlers]]
async = false
file_path = '/var/log/3fs/mgmtd_main-fatal.log'
max_file_size = '100MB'
max_files = 10
name = 'fatal'
rotate = true
rotate_on_open = false
start_level = 'FATAL'
stream_type = 'STDERR'
writer_type = 'STREAM'

[common.memory]
prof_active = false
prof_prefix = ''

[common.monitor]
collect_period = '1s'
num_collectors = 1

[[common.monitor.reporters]]
type = 'monitor_collector'

[common.monitor.reporters.monitor_collector]
remote_ip = ""

[server.base.independent_thread_pool]
bg_thread_pool_stratetry = 'SHARED_QUEUE'
collect_stats = false
enable_work_stealing = false
io_thread_pool_stratetry = 'SHARED_QUEUE'
num_bg_threads = 2
num_connect_threads = 2
num_io_threads = 2
num_proc_threads = 2
proc_thread_pool_stratetry = 'SHARED_QUEUE'

[server.base.thread_pool]
bg_thread_pool_stratetry = 'SHARED_QUEUE'
collect_stats = false
enable_work_stealing = false
io_thread_pool_stratetry = 'SHARED_QUEUE'
num_bg_threads = 2
num_connect_threads = 2
num_io_threads = 2
num_proc_threads = 2
proc_thread_pool_stratetry = 'SHARED_QUEUE'

[[server.base.groups]]
check_connections_interval = '1min'
connection_expiration_time = '1day'
network_type = 'RDMA'
services = [ 'Mgmtd' ]
use_independent_thread_pool = false

[server.base.groups.io_worker]
num_event_loop = 1
rdma_connect_timeout = '5s'
read_write_rdma_in_event_thread = false
read_write_tcp_in_event_thread = false
tcp_connect_timeout = '1s'
wait_to_retry_send = '100ms'

[server.base.groups.io_worker.connect_concurrency_limiter]
max_concurrency = 4

[server.base.groups.io_worker.ibsocket]
buf_ack_batch = 8
buf_signal_batch = 8
buf_size = 16384
drain_timeout = '5s'
drop_connections = 0
event_ack_batch = 128
max_rd_atomic = 16
max_rdma_wr = 128
max_rdma_wr_per_post = 32
max_sge = 16
min_rnr_timer = 1
record_bytes_per_peer = false
record_latency_per_peer = false
retry_cnt = 7
rnr_retry = 0
send_buf_cnt = 32
sl = 0
start_psn = 0
timeout = 14

[server.base.groups.io_worker.transport_pool]
max_connections = 1

[server.base.groups.listener]
domain_socket_index = 1
filter_list = []
listen_port = 8000
listen_queue_depth = 4096
rdma_accept_timeout = '15s'
rdma_listen_ethernet = true
reuse_port = false

[server.base.groups.processor]
enable_coroutines_pool = true
max_coroutines_num = 256
max_processing_requests_num = 4096
response_compression_level = 1
response_compression_threshold = '128KB'

[[server.base.groups]]
check_connections_interval = '1min'
connection_expiration_time = '1day'
network_type = 'TCP'
services = [ 'Core' ]
use_independent_thread_pool = true

[server.base.groups.io_worker]
num_event_loop = 1
rdma_connect_timeout = '5s'
read_write_rdma_in_event_thread = false
read_write_tcp_in_event_thread = false
tcp_connect_timeout = '1s'
wait_to_retry_send = '100ms'

[server.base.groups.io_worker.connect_concurrency_limiter]
max_concurrency = 4

[server.base.groups.io_worker.ibsocket]
buf_ack_batch = 8
buf_signal_batch = 8
buf_size = 16384
drain_timeout = '5s'
drop_connections = 0
event_ack_batch = 128
max_rd_atomic = 16
max_rdma_wr = 128
max_rdma_wr_per_post = 32
max_sge = 16
min_rnr_timer = 1
record_bytes_per_peer = false
record_latency_per_peer = false
retry_cnt = 7
rnr_retry = 0
send_buf_cnt = 32
sl = 0
start_psn = 0
timeout = 14

[server.base.groups.io_worker.transport_pool]
max_connections = 1

[server.base.groups.listener]
domain_socket_index = 1
filter_list = []
listen_port = 9000
listen_queue_depth = 4096
rdma_accept_timeout = '15s'
rdma_listen_ethernet = true
reuse_port = false

[server.base.groups.processor]
enable_coroutines_pool = true
max_coroutines_num = 256
max_processing_requests_num = 4096
response_compression_level = 1
response_compression_threshold = '128KB'

[server.service]
allow_heartbeat_from_unregistered = true
authenticate = false
bootstrapping_length = '2min'
bump_routing_info_version_interval = '5s'
check_status_interval = '10s'
client_session_timeout = '20min'
enable_routinginfo_cache = true
extend_lease_check_release_version = true
extend_lease_interval = '10s'
heartbeat_fail_interval = '1min'
heartbeat_ignore_stale_targets = true
heartbeat_ignore_unknown_targets = false
heartbeat_timestamp_valid_window = '30s'
lease_length = '1min'
new_chain_bootstrap_interval = '2min'
only_accept_client_uuid = false
retry_times_on_txn_errors = -1
send_heartbeat = true
send_heartbeat_interval = '10s'
suspicious_lease_interval = '20s'
target_info_load_interval = '1s'
target_info_persist_batch = 1000
target_info_persist_interval = '1s'
try_adjust_target_order_as_preferred = false
update_chains_interval = '1s'
update_metrics_interval = '1s'
validate_lease_on_write = true

[server.service.retry_transaction]
max_backoff = '1s'
max_retry_count = 10

[server.service.user_cache]
buckets = 127
exist_ttl = '5min'
inexist_ttl = '10s'`
	return mainApp, mainLauncher, mainContent
}

func (s *prepareMgmtdConfigStepSuite) testPrepareConfig(removeAllErr error) {
	s.MockedLocal.On("MkdirAll", s.Runtime.Services.Mgmtd.WorkDir).Return(nil)
	tmpDir := "/root/tmp..."
	s.MockedLocal.On("MkdirTemp", s.Runtime.Services.Mgmtd.WorkDir, s.node.Name+".*").Return(tmpDir, nil)
	s.MockedLocal.On("RemoveAll", tmpDir).Return(removeAllErr)
	mainAppConfig, mainLauncherConfig, mainConfig := s.getGeneratedConfigContent()
	s.mockGenConfig(tmpDir+"/mgmtd_main_app.toml", mainAppConfig)
	s.mockGenConfig(tmpDir+"/mgmtd_main_launcher.toml", mainLauncherConfig)
	s.mockGenConfig(tmpDir+"/mgmtd_main.toml", mainConfig)
	s.MockedLocal.On("WriteFile", tmpDir+"/fdb.cluster", []byte(s.fdbContent), os.FileMode(0644)).Return(nil)
	s.MockRunner.On("Scp", tmpDir, s.Cfg.Services.Mgmtd.WorkDir).Return(nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockedLocal.AssertExpectations(s.T())
	s.MockOS.AssertExpectations(s.T())
	s.MockRunner.AssertExpectations(s.T())
}

func (s *prepareMgmtdConfigStepSuite) TestPrepareConfig() {
	s.testPrepareConfig(nil)
}

func (s *prepareMgmtdConfigStepSuite) TestPrepareConfigWithRemoveTempDirFailed() {
	s.testPrepareConfig(errors.New("remove temp dir failed"))
}
