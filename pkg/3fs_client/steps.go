package fsclient

import (
	"context"
	"strings"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/task"
)

type umountHostMountponitStep struct {
	task.BaseStep
}

func (s *umountHostMountponitStep) Execute(ctx context.Context) error {
	mp := s.Runtime.Services.Client.HostMountpoint

	out, err := s.Em.Runner.Exec(ctx, "mount")
	if err != nil {
		return errors.Annotate(err, "get mountponits")
	}
	if !strings.Contains(out, mp) {
		s.Logger.Infof("%s is not mounted, skip umount it", mp)
		return nil
	}

	_, err = s.Em.Runner.Exec(ctx, "umount", s.Runtime.Services.Client.HostMountpoint)
	if err != nil {
		return errors.Annotatef(err, "umount %s", mp)
	}
	s.Logger.Infof("Successfully umount %s", mp)
	return nil
}
