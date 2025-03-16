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

package config

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/tests/base"
)

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(imageSuite))
}

type imageSuite struct {
	base.Suite
}

func (s *imageSuite) TestGetImage() {
	cfg := NewConfigWithDefaults()
	cfg.Images.FFFS.Tag = "1.1.1"

	img, err := cfg.Images.GetImage(ImageName3FS)
	s.NoError(err)
	s.Equal("open3fs/3fs:1.1.1", img)
}

func (s *imageSuite) TestGetImageWithRegistry() {
	cfg := NewConfigWithDefaults()
	cfg.Images.Registry = "hub.docker.com"
	cfg.Images.FFFS.Tag = "1.1.1"

	img, err := cfg.Images.GetImage(ImageName3FS)
	s.NoError(err)
	s.Equal("hub.docker.com/open3fs/3fs:1.1.1", img)
}
