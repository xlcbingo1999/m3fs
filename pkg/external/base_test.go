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

package external_test

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/log"
)

var suiteRun = suite.Run

type Suite struct {
	suite.Suite

	r  *MockedRunner
	em *external.Manager
}

func (s *Suite) SetupSuite() {
	s.T().Parallel()
}

func (s *Suite) SetupTest() {
	s.r = NewMockedRunner(s.T())
	log.InitLogger(logrus.DebugLevel)
	s.em = external.NewManager(s.r, log.Logger)
}

// Ctx returns a context used in test.
func (s *Suite) Ctx() context.Context {
	return context.TODO()
}

// R returns a require context.
func (s *Suite) R() *require.Assertions {
	return s.Require()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
