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
