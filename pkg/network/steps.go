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

package network

import (
	"context"
	"path"
	"strings"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/task"
)

const (
	ibdev2netdevScript = `#!/bin/bash

# emulate mlnx util ibdev2netdev
# format: ibdev port xx ==> netdev (Up/Down)
# example: mlx5_0 port 1 ==> eth0 (Up)

rdma link | awk '{print $2,$4,$8}' | while read ibInfo state netdev
do
	ibdev=$(echo $ibInfo | cut -d/ -f1)
	ibport=$(echo $ibInfo | cut -d/ -f2)

    if [ "$state" = "ACTIVE" ]; then
        state=Up
    else
        state=Down
    fi
    if [ -n "$netdev" ]; then
        echo "$ibdev port $ibport ==> $netdev ($state)"
    fi
done
`
	createRdmaLinkScript = `#!/bin/bash

for netdev in $(ip -o -4 a | awk '{print $2}' | grep -vw lo | sort -u)
do
    # skip linux bridge
    if ip -o -d l show $netdev | grep -q bridge_id; then
        continue
    fi

    if rdma link | grep -q -w "netdev $netdev"; then
        continue
    fi

    echo "Create rdma link for $netdev"
    rxe_name="${netdev}_rxe0"
    rdma link add $rxe_name type rxe netdev $netdev
    if rdma link | grep -q -w "link $rxe_name"; then
        echo "Success to create $rxe_name"
    fi
done
`
)

type genIbdev2netdevScriptStep struct {
	task.BaseStep
}

func (s *genIbdev2netdevScriptStep) Execute(ctx context.Context) error {
	s.Logger.Debugf("Generating ibdev2netdev script for %s", s.Node.Host)
	localEm := s.Runtime.LocalEm
	tmpDir, err := localEm.FS.MkdirTemp("/tmp", "m3fs-prepare-network.*")
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := localEm.FS.RemoveAll(tmpDir); err != nil {
			s.Logger.Errorf("Failed to remove temporary directory %s: %v", tmpDir, err)
		}
	}()
	scriptLocalPath := path.Join(tmpDir, "ibdev2netdev")
	err = localEm.FS.WriteFile(scriptLocalPath, []byte(ibdev2netdevScript), 0755)
	if err != nil {
		return errors.Trace(err)
	}

	binDir := path.Join(s.Runtime.Cfg.WorkDir, "bin")
	_, err = s.Em.Runner.Exec(ctx, "mkdir", "-p", binDir)
	if err != nil {
		return errors.Annotatef(err, "mkdir %s", binDir)
	}
	remotePath := path.Join(binDir, "ibdev2netdev")
	if err := s.Em.Runner.Scp(scriptLocalPath, remotePath); err != nil {
		return errors.Annotatef(err, "scp %s", scriptLocalPath)
	}
	_, err = s.Em.Runner.Exec(ctx, "chmod", "+x", remotePath)
	if err != nil {
		return errors.Annotatef(err, "chmod +x %s", remotePath)
	}

	return nil
}

type loadRdmaRxeModuleStep struct {
	task.BaseStep
}

func (s *loadRdmaRxeModuleStep) Execute(ctx context.Context) error {
	s.Logger.Debugf("Loading rdma_rxe kernel module for %s if needed", s.Node.Host)

	output, err := s.Em.Runner.Exec(ctx, "ls", "/sys/module")
	if err != nil {
		return errors.Annotate(err, "")
	}
	kernelModules := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		modules := strings.Fields(line)
		for _, module := range modules {
			kernelModules[module] = struct{}{}
		}
	}

	// if any of those modules is loaded, we don't need to load rdma_rxe
	for _, module := range []string{"mlx5_core", "irdma", "erdma", "rdma_rxe"} {
		if _, ok := kernelModules[module]; ok {
			return nil
		}
	}

	_, err = s.Em.Runner.Exec(ctx, "modprobe", "rdma_rxe")
	if err != nil {
		return errors.Annotatef(err, "modprobe rdma_rxe")
	}
	return nil
}

type createRdmaRxeLinkStep struct {
	task.BaseStep
}

func (s *createRdmaRxeLinkStep) Execute(ctx context.Context) error {
	s.Logger.Debugf("Creating rdma link for %s", s.Node.Host)
	localEm := s.Runtime.LocalEm
	tmpDir, err := localEm.FS.MkdirTemp("/tmp", "m3fs-prepare-network.*")
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := localEm.FS.RemoveAll(tmpDir); err != nil {
			s.Logger.Errorf("Failed to remove temporary directory %s: %v", tmpDir, err)
		}
	}()

	scriptLocalPath := path.Join(tmpDir, "create_rdma_rxe_link")
	err = localEm.FS.WriteFile(scriptLocalPath, []byte(createRdmaLinkScript), 0755)
	if err != nil {
		return errors.Trace(err)
	}

	binDir := path.Join(s.Runtime.Cfg.WorkDir, "bin")
	_, err = s.Em.Runner.Exec(ctx, "mkdir", "-p", binDir)
	if err != nil {
		return errors.Annotatef(err, "mkdir %s", binDir)
	}
	remotePath := path.Join(binDir, "create_rdma_rxe_link")
	if err := s.Em.Runner.Scp(scriptLocalPath, remotePath); err != nil {
		return errors.Annotatef(err, "scp %s", scriptLocalPath)
	}
	_, err = s.Em.Runner.Exec(ctx, "chmod", "+x", remotePath)
	if err != nil {
		return errors.Annotatef(err, "chmod +x %s", remotePath)
	}
	_, err = s.Em.Runner.Exec(ctx, "bash", remotePath)
	if err != nil {
		return errors.Annotatef(err, "execute %s", remotePath)
	}
	return nil
}
