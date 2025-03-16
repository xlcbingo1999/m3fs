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
	"fmt"
	"net/url"

	"github.com/open3fs/m3fs/pkg/errors"
)

// defines image names
const (
	ImageNameFdb        = "fdb"
	ImageNameClickhouse = "clickhouse"
	ImageName3FS        = "3fs"
)

// Image is component container image config
type Image struct {
	Repo string
	Tag  string
}

// Images contains all component container image configs
type Images struct {
	Registry   string `yaml:"registry"`
	FFFS       Image  `yaml:"3fs"` // 3fs cannot used as struct filed name, so we use fffs instead
	Clickhouse Image  `yaml:"clickhouse"`
	Fdb        Image  `yaml:"fdb"`
}

// GetImage get image path of target component
func (i *Images) GetImage(imgName string) (string, error) {
	var img Image
	switch imgName {
	case ImageNameFdb:
		img = i.Fdb
	case ImageName3FS:
		img = i.FFFS
	case ImageNameClickhouse:
		img = i.Clickhouse
	default:
		return "", errors.Errorf("invalid image name %s", imgName)
	}

	imagePath := fmt.Sprintf("%s:%s", img.Repo, img.Tag)
	if i.Registry != "" {
		var err error
		imagePath, err = url.JoinPath(i.Registry, imagePath)
		if err != nil {
			return "", errors.Annotatef(err, "get image path of %s", imgName)
		}
	}

	return imagePath, nil
}
