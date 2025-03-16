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
	"github.com/open3fs/m3fs/pkg/task"
)

// PrepareNetworkTask is a task for preparing network for a new node.
type PrepareNetworkTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *PrepareNetworkTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("PrepareNetworkTask")
	nodes := r.Cfg.Nodes

	t.SetSteps([]task.StepConfig{
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(genIbdev2netdevScriptStep) },
		},
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(loadRdmaRxeModuleStep) },
		},
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(createRdmaRxeLinkStep) },
		},
	})
}
