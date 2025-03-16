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

package task

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/tests/base"
)

type baseSuite struct {
	base.Suite
}

var suiteRun = suite.Run

type mockTask struct {
	mock.Mock
	Interface
}

func (m *mockTask) Run(context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTask) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockTask) Init(r *Runtime) {
	m.Called(r)
}
