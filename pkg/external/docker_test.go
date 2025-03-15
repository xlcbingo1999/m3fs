package external_test

import (
	"testing"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/external"
)

func TestDockerRunSuite(t *testing.T) {
	suiteRun(t, new(dockerRunSuite))
}

type dockerRunSuite struct {
	Suite
}

func (s *dockerRunSuite) Test() {
	containerName := "3fs-clickhouse"
	detach := true
	hostAddress := "127.0.0.1"
	protocol := "tcp"
	args := &external.RunArgs{
		Image:      "clickhouse/clickhouse-server:latest",
		Name:       &containerName,
		Detach:     &detach,
		Entrypoint: common.Pointer("''"),
		Rm:         common.Pointer(true),
		Command:    []string{"ls"},
		Privileged: common.Pointer(true),
		Ulimits: map[string]string{
			"nproc": "65535:65535",
		},
		Envs: map[string]string{
			"A": "B",
		},
		HostNetwork: true,
		Publish: []*external.PublishArgs{
			{
				HostAddress:   &hostAddress,
				HostPort:      9000,
				ContainerPort: 9000,
				Protocol:      &protocol,
			},
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: "/path/to/data",
				Target: "/clickhouse/data",
				Rshare: common.Pointer(true),
			},
		},
	}
	mockCmd := "docker run --name 3fs-clickhouse --detach --network host -e A=B --entrypoint '' --rm " +
		"--privileged --ulimit nproc=65535:65535 -p 127.0.0.1:9000:9000/tcp " +
		"--volume /path/to/data:/clickhouse/data:rshared clickhouse/clickhouse-server:latest ls"
	s.r.MockExec(mockCmd, "", nil)
	_, err := s.em.Docker.Run(s.Ctx(), args)
	s.NoError(err)
}

func TestDockerRmSuite(t *testing.T) {
	suiteRun(t, new(dockerRmSuite))
}

type dockerRmSuite struct {
	Suite
}

func (s *dockerRmSuite) Test() {
	mockCmd := "docker rm --force test"
	s.r.MockExec(mockCmd, "", nil)
	_, err := s.em.Docker.Rm(s.Ctx(), "test", true)
	s.NoError(err)
}

func TestDockerExecSuite(t *testing.T) {
	suiteRun(t, new(dockerExecSuite))
}

type dockerExecSuite struct {
	Suite
}

func (s *dockerExecSuite) Test() {
	mockCmd := "docker exec fdb fdbcli --exec status"
	s.r.MockExec(mockCmd, "", nil)
	_, err := s.em.Docker.Exec(s.Ctx(), "fdb", "fdbcli", "--exec", "status")
	s.NoError(err)
}
