/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"archive/tar"
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/snapshot"
	"github.com/containers/image/docker"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// CreateFile creates a file at path with contents specified
func CreateFile(path string, contents []byte) error {
	// Create directory path if it doesn't exist
	baseDir := filepath.Dir(path)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		logrus.Debugf("baseDir %s for file %s does not exist. Creating.", baseDir, path)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(contents)
	return err
}

// Files returns a list of all files that stem from root
func Files(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return err
	})
	return files, err
}

// IsDir checks if path is a directory
func IsDir(path string) (bool, error) {
	f, err := os.Stat(path)
	return f.IsDir(), err
}

func ExtractFileSystemFromImage(img string) error {
	ref, err := docker.ParseReference("//" + img)
	if err != nil {
		return err
	}
	imgSrc, err := ref.NewImageSource(nil)
	if err != nil {
		return err
	}
	return pkgutil.GetFileSystemFromReference(ref, imgSrc, constants.RootDir, constants.Whitelist)
}

func GetImageTar(name string) (string, error) {
	filepath := filepath.Join(constants.WorkDir, name+".tar")
	_, err := os.Stat(filepath)
	return filepath, err
}

func SaveFileSystemAsTarball(dest string) error {
	tarPath := filepath.Join(constants.WorkDir, dest+".tar")
	f, err := os.Create(tarPath)
	logrus.Infof("Created tarball to save filesystem in at %s", tarPath)
	defer f.Close()
	if err != nil {
		return err
	}
	w := tar.NewWriter(f)
	defer w.Close()

	err = filepath.Walk(constants.RootDir, func(path string, info os.FileInfo, err error) error {
		if snapshot.IgnorePath(path, constants.RootDir) {
			return nil
		}
		return snapshot.AddToTar(path, info, w)
	})
	return err
}

func DeleteFileSystem() error {
	logrus.Info("Deleting filesystem...")
	err := filepath.Walk(constants.RootDir, func(path string, info os.FileInfo, err error) error {
		if snapshot.IgnorePath(path, constants.RootDir) || path == constants.RootDir {
			return nil
		}
		logrus.Debugf("Deleting %s", path)
		e := os.RemoveAll(path)
		if e != nil {
			logrus.Debugf("Couldn't remove %s: %s", path, e)
		}
		return nil
	})
	return err
}
