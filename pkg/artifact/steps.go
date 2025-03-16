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

package artifact

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/task"
)

type prepareTmpDirStep struct {
	task.BaseLocalStep
}

func (s *prepareTmpDirStep) Execute(context.Context) error {
	tmpDir, ok := s.Runtime.LoadString(task.RuntimeArtifactTmpDirKey)
	if !ok {
		return errors.Errorf("Failed to get value of %s", task.RuntimeArtifactTmpDirKey)
	}
	if err := s.Runtime.LocalEm.FS.MkdirAll(tmpDir); err != nil {
		return errors.Trace(err)
	}
	return nil
}

type downloadImagesStep struct {
	task.BaseLocalStep
}

func (s *downloadImagesStep) Execute(context.Context) error {
	imageNames := []string{
		config.ImageNameFdb,
		config.ImageNameClickhouse,
		config.ImageName3FS,
	}
	for _, imageName := range imageNames {
		filePath, err := s.downloadImage(imageName)
		if err != nil {
			return errors.Trace(err)
		}
		var filePaths []string
		if filePathsValue, ok := s.Runtime.Load(task.RuntimeArtifactFilePathsKey); ok {
			filePaths = filePathsValue.([]string)
		}
		filePaths = append(filePaths, filePath)
		s.Runtime.Store(task.RuntimeArtifactFilePathsKey, filePaths)
	}
	return nil
}

func (s *downloadImagesStep) getUrl(fileName string) string {
	return fmt.Sprintf("https://artifactory.open3fs.com/3fs/%s", fileName)
}

func (s *downloadImagesStep) downloadImage(imageName string) (string, error) {
	imageFileName, err := s.Runtime.Cfg.Images.GetImageFileName(imageName)
	if err != nil {
		return "", errors.Trace(err)
	}
	imageUrl := s.getUrl(imageFileName)
	imageSumFileName := fmt.Sprintf("%s.sha256sum", imageFileName)
	imageSumUrl := s.getUrl(imageSumFileName)

	tmpDir, ok := s.Runtime.LoadString(task.RuntimeArtifactTmpDirKey)
	if !ok {
		return "", errors.Errorf("Failed to get tmp dir for artifact")
	}
	dstPath := filepath.Join(tmpDir, imageFileName)
	notExisted, err := s.Runtime.LocalEm.FS.IsNotExist(dstPath)
	if err != nil {
		return "", errors.Trace(err)
	}
	if !notExisted {
		s.Logger.Infof("File of %s image exists", imageName)
		sumContent, err := s.Runtime.LocalEm.FS.ReadRemoteFile(imageSumUrl)
		if err != nil {
			return "", errors.Trace(err)
		}
		expectedSum := strings.Split(sumContent, " ")[0]
		actualSum, err := s.Runtime.LocalEm.FS.Sha256sum(dstPath)
		if err != nil {
			return "", errors.Trace(err)
		}
		if expectedSum == actualSum {
			s.Logger.Infof("Skip downloading existed %s image", imageName)
			return dstPath, nil
		}
		s.Logger.Infof("Current sha256sum of file %s is %s, expected %s",
			dstPath, actualSum, expectedSum)
	}

	s.Logger.Infof("Downloading %s image from %s", imageName, imageUrl)
	if err := s.Runtime.LocalEm.FS.DownloadFile(imageUrl, dstPath); err != nil {
		return "", errors.Trace(err)
	}
	s.Logger.Infof("Downloaded %s image", imageName)

	return dstPath, nil
}

type tarFilesStep struct {
	task.BaseLocalStep
}

func (s *tarFilesStep) Execute(context.Context) error {
	filePathsValue, ok := s.Runtime.Load(task.RuntimeArtifactFilePathsKey)
	if !ok {
		return errors.Errorf("Failed to get value of %s", task.RuntimeArtifactFilePathsKey)
	}
	filePaths := filePathsValue.([]string)
	dstPath, ok := s.Runtime.LoadString(task.RuntimeArtifactOutputPathKey)
	if !ok {
		return errors.Errorf("Failed to get value of %s", task.RuntimeArtifactOutputPathKey)
	}
	tmpDir, ok := s.Runtime.LoadString(task.RuntimeArtifactTmpDirKey)
	if !ok {
		return errors.Errorf("Failed to get value of %s", task.RuntimeArtifactTmpDirKey)
	}

	s.Logger.Infof("Generating tar files %s", dstPath)
	if err := s.Runtime.LocalEm.FS.Tar(filePaths, tmpDir, dstPath); err != nil {
		return errors.Trace(err)
	}
	s.Logger.Infof("Generated tar files %s", dstPath)
	return nil
}
