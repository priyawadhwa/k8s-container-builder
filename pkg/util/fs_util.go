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
	"bufio"
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/containers/image/docker"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var whitelist = []string{"/kbuild"}

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
	whitelist, err := fileSystemWhitelist(constants.WhitelistPath)
	if err != nil {
		return err
	}
	logrus.Infof("Whitelisted directories are %s", whitelist)
	return pkgutil.GetFileSystemFromReference(ref, imgSrc, constants.RootDir, whitelist)
}

// PathInWhitelist returns true if the path is whitelisted
func PathInWhitelist(path, directory string) bool {
	for _, d := range whitelist {
		dirPath := filepath.Join(directory, d)
		if pkgutil.HasFilepathPrefix(path, dirPath) {
			return true
		}
	}
	return false
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

// Files returns a list of all files at the filepath relative to root
func Files(fp string, root string) ([]string, error) {
	var files []string
	fullPath := filepath.Join(root, fp)
	logrus.Debugf("Getting files and contents at root %s", fullPath)
	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, relPath)
		return err
	})
	return files, err
}

// FilepathExists returns true if the path exists
func FilepathExists(path string) bool {
	_, err := os.Stat(path)
	return (err == nil)
}

// CreateFile creates a file at path with contents specified
func CreateFile(path string, contents []byte, perm os.FileMode) error {
	// Create directory path if it doesn't exist
	baseDir := filepath.Dir(path)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		logrus.Debugf("baseDir %s for file %s does not exist. Creating.", baseDir, path)
		if err := os.MkdirAll(baseDir, perm); err != nil {
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

// GetMatchedFiles returns a map of [src]:[matching filenames], used to resolve wildcards
func GetMatchedFiles(srcs, files []string) (map[string][]string, error) {
	f := make(map[string][]string)
	for _, src := range srcs {
		src = filepath.Clean(src)
		matchedFiles := []string{}
		for _, file := range files {
			matched, err := filepath.Match(src, file)
			if err != nil {
				return nil, err
			}
			matchedRoot, err := filepath.Match(filepath.Join(constants.RootDir, src), file)
			if err != nil {
				return nil, err
			}
			keep := matched || matchedRoot || pkgutil.HasFilepathPrefix(file, src)
			if !keep {
				continue
			}
			matchedFiles = append(matchedFiles, file)
		}
		f[src] = matchedFiles
	}
	return f, nil
}
