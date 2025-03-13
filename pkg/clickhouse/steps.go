package clickhouse

import (
	"context"
	"path"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
)

type startContainerStep struct {
	task.BaseStep
}

func (s *startContainerStep) Execute(ctx context.Context) error {
	workDir := s.Runtime.Services.Clickhouse.WorkDir
	dataDir := path.Join(workDir, "data")
	logDir := path.Join(workDir, "log")
	configDir := path.Join(workDir, "config.d")
	sqlDir := path.Join(workDir, "sql")
	_, err := s.Em.OS.Exec(ctx, "mkdir", "-p", dataDir)
	if err != nil {
		return errors.Annotatef(err, "mkdir %s", dataDir)
	}
	_, err = s.Em.OS.Exec(ctx, "mkdir", "-p", logDir)
	if err != nil {
		return errors.Annotatef(err, "mkdir %s", dataDir)
	}
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "clickhouse")
	if err != nil {
		return errors.Trace(err)
	}
	args := &external.RunArgs{
		Image:       img,
		Name:        &s.Runtime.Services.Clickhouse.ContainerName,
		HostNetwork: true,
		Envs: map[string]string{
			"CLICKHOUSE_DB":       s.Runtime.Services.Clickhouse.Db,
			"CLICKHOUSE_USER":     s.Runtime.Services.Clickhouse.User,
			"CLICKHOUSE_PASSWORD": s.Runtime.Services.Clickhouse.Password,
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/lib/clickhouse",
			},
			{
				Source: logDir,
				Target: "/var/log/clickhouse-server",
			},
			{
				Source: configDir,
				Target: "/etc/clickhouse-server/config.d",
			},
			{
				Source: sqlDir,
				Target: "/tmp/sql",
			},
		},
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	s.Logger.Infof("Started container %s", s.Runtime.Services.Clickhouse.ContainerName)
	return nil
}
