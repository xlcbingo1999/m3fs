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

package image

import (
	"fmt"
	"net/url"

	"github.com/open3fs/m3fs/pkg/errors"
)

// Image is the Image info definition
type Image struct {
	Repo string
	Tag  string
}

// GetURL returns full url of image
func (i *Image) GetURL(registry string) (string, error) {
	imgUrl := fmt.Sprintf("%s:%s", i.Repo, i.Tag)
	if registry != "" {
		var err error
		imgUrl, err = url.JoinPath(registry, imgUrl)
		if err != nil {
			return "", errors.Annotatef(err, "get %s url", i.Repo)
		}
	}

	return imgUrl, nil
}

var images = map[string]Image{
	"3fs":        {Repo: "open3fs/3fs", Tag: "20250315"},
	"fdb":        {Repo: "open3fs/foundationdb", Tag: "7.3.63"},
	"clickhouse": {Repo: "open3fs/clickhouse", Tag: "25.1-jammy"},
}

// GetImage returns image full url by component name
func GetImage(registry, compo string) (string, error) {
	i, ok := images[compo]
	if !ok {
		return "", errors.Errorf("image of %s not found", compo)
	}

	return i.GetURL(registry)
}
