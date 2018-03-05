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
	"bufio"
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/containers/image/docker"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var whitelist = []string{"/workspace"}

func InitializeWhitelist() error {
	whitelist, err := fileSystemWhitelist(constants.WhitelistPath)
	logrus.Infof("Whitelisted directories are %s", whitelist)
	return err
}

// ExtractFileSystemFromImage pulls an image and unpacks it to a file system at root
func ExtractFileSystemFromImage(img string) error {
	ref, err := docker.ParseReference("//" + img)
	if err != nil {
		return err
	}
	imgSrc, err := ref.NewImageSource(nil)
	if err != nil {
		return err
	}
	return pkgutil.GetFileSystemFromReference(ref, imgSrc, constants.RootDir, whitelist)
}

// Get whitelist from roots of mounted files
// Each line of /proc/self/mountinfo is in the form:
// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)
// Where (5) is the mount point relative to the process's root
// From: https://www.kernel.org/doc/Documentation/filesystems/proc.txt
func fileSystemWhitelist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		logrus.Debugf("Read the following line from %s: %s", path, line)
		if err != nil && err != io.EOF {
			return nil, err
		}
		lineArr := strings.Split(line, " ")
		if len(lineArr) < 5 {
			if err == io.EOF {
				logrus.Debugf("Reached end of file %s", path)
				break
			}
			continue
		}
		if lineArr[4] != constants.RootDir {
			logrus.Debugf("Appending %s from line: %s", lineArr[4], line)
			whitelist = append(whitelist, lineArr[4])
		}
		if err == io.EOF {
			logrus.Debugf("Reached end of file %s", path)
			break
		}
	}
	return whitelist, nil
}

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

// FilesAndContents returns a map of filename:file contents for all files that stem from root
// The filepath is relative to root
func FilesAndContents(fp string, root string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	logrus.Debugf("Getting files and contents at root %s", root)
	err := filepath.Walk(fp, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		logrus.Debugf("Reading file %s", path)
		contents, e := ioutil.ReadFile(path)
		if e != nil {
			return e
		}
		relPath, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		logrus.Debugf("Adding file %s to map of files", relPath)
		files[relPath] = contents
		return err
	})
	return files, err
}

// IsDir checks if path is a directory
func IsDir(path string) (bool, error) {
	f, err := os.Stat(path)
	return f.IsDir(), err
}

func FilepathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetImageTar(from string) (string, error) {
	tarPath := filepath.Join(constants.WorkspaceDir, from+".tar")
	if _, err := os.Stat(tarPath); err != nil {
		return "", err
	}
	return tarPath, nil
}

func SaveFileSystemAsTarball(name string, index int) error {
	tarPath := filepath.Join(constants.WorkspaceDir, name+".tar")
	if name == "" {
		tarPath = filepath.Join(constants.WorkspaceDir, strconv.Itoa(index)+".tar")
	}
	f, err := os.Create(tarPath)
	logrus.Infof("Created tarball to save filesystem in at %s", tarPath)
	defer f.Close()
	if err != nil {
		return err
	}
	w := tar.NewWriter(f)
	defer w.Close()

	err = filepath.Walk(constants.RootDir, func(path string, info os.FileInfo, err error) error {
		if IgnoreFilepath(path, constants.RootDir) {
			return nil
		}
		if strings.Contains(path, "pkg/foo") {
			logrus.Infof("################# %s", path)
		}
		return AddToTar(path, info, w)
	})
	if err != nil {
		return err
	}

	// Symlink
	indexPath := filepath.Join(constants.WorkspaceDir, strconv.Itoa(index)+".tar")
	if indexPath != tarPath {
		logrus.Debugf("Symlinking from %s to %s", tarPath, indexPath)
		return os.Symlink(tarPath, indexPath)
	}
	return nil
}

func DeleteFileSystem() error {
	logrus.Info("Deleting filesystem...")
	err := filepath.Walk(constants.RootDir, func(path string, info os.FileInfo, err error) error {
		if IgnoreFilepathForDeletion(path, constants.RootDir) || path == constants.RootDir {
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

// TODO: ignore anything in /proc/self/mounts
// ignore anything in the whitelist
func IgnoreFilepath(p, directory string) bool {
	for _, d := range whitelist {
		dirPath := filepath.Join(directory, d)
		if pkgutil.HasFilepathPrefix(p, dirPath) {
			return true
		}
	}
	return false
}

func IgnoreFilepathForDeletion(p, directory string) bool {
	deleteWhitelist := append(whitelist, constants.ConfigPath)
	for _, d := range deleteWhitelist {
		dirPath := filepath.Join(directory, d)
		if filepath.Clean(dirPath) == filepath.Clean(p) {
			return true
		}
		if pkgutil.HasFilepathPrefix(dirPath, p) || pkgutil.HasFilepathPrefix(p, dirPath) {
			return true
		}
	}
	return false
}
