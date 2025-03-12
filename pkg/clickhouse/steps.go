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
	dataDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "data")
	_, err := s.Em.OS.Exec(ctx, "mkdir", "-p", dataDir)
	if err != nil {
		return errors.Annotatef(err, "mkdir %s", dataDir)
	}
	logDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "log")
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
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/lib/clickhouse",
			},
			{
				Source: logDir,
				Target: "/var/log/clickhouse-server",
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
