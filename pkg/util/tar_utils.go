/*
Copyright 2017 Google, Inc. All rights reserved.

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
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func unpackTar(tr *tar.Reader, path string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			logrus.Error("Error getting next tar header")
			return err
		}
		if strings.Contains(header.Name, ".wh.") {
			rmPath := filepath.Join(path, header.Name)
			// Remove the .wh file if it was extracted.
			if _, err := os.Stat(rmPath); !os.IsNotExist(err) {
				if err := os.Remove(rmPath); err != nil {
					logrus.Error(err)
				}
			}

			// Remove the whited-out path.
			newName := strings.Replace(rmPath, ".wh.", "", 1)
			if err = walkAndRemove(newName); err != nil {
				logrus.Error(err)
			}
			continue
		}

		target := filepath.Join(path, header.Name)
		mode := header.FileInfo().Mode()

		if checkWhitelist(target) {
			continue
		}
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, mode); err != nil {
					return err
				}
				if err := os.Chmod(target, mode); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			// It's possible for a file to be included before the directory it's in is created.
			baseDir := filepath.Dir(target)
			if _, err := os.Stat(baseDir); os.IsNotExist(err) {
				logrus.Debugf("baseDir %s for file %s does not exist. Creating.", baseDir, target)
				if err := os.MkdirAll(baseDir, 0755); err != nil {
					return err
				}
			}
			// It's possible we end up creating files that can't be overwritten based on their permissions.
			// Explicitly delete an existing file before continuing.
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				logrus.Debugf("Removing %s for overwrite.", target)
				if err := os.Remove(target); err != nil {
					return err
				}
			}

			currFile, err := os.Create(target)
			if err != nil {
				logrus.Errorf("Error creating file %s %s", target, err)
				return err
			}
			// manually set permissions on file, since the default umask (022) will interfere
			if err = os.Chmod(target, mode); err != nil {
				logrus.Errorf("Error updating file permissions on %s", target)
				return err
			}
			_, err = io.Copy(currFile, tr)
			if err != nil {
				return err
			}
			currFile.Close()
		case tar.TypeSymlink:
			// It's possible we end up creating files that can't be overwritten based on their permissions.
			// Explicitly delete an existing file before continuing.
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				logrus.Debugf("Removing %s to create symlink.", target)
				walkAndRemove(target)
			}

			err = os.Symlink(header.Linkname, target)
			if err != nil {
				logrus.Errorf("Failed to create symlink between %s and %s: %s", header.Linkname, target, err)
			}
		}

	}
	return nil
}

func walkAndRemove(p string) error {
	filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err = os.Chmod(path, 0777); err != nil {
			logrus.Errorf("Error updating file permissions on %s before removing for symlink creation", path)
			return err
		}
		if err := os.RemoveAll(path); err != nil {
			logrus.Errorf("Failed to delete %s, and it's contents: %s", path, err)
		}
		return nil
	})
	return nil
}

func checkWhiteouts(file string) bool {
	// Check if file is in whiteouts
	if _, ok := whiteouts[file]; ok {
		if whiteouts[file] {
			return true
		}
	}
	// Check if file is in a whiteout directory
	for {
		directory := filepath.Dir(file)
		if directory == file {
			break
		}
		if _, ok := whiteouts[directory]; ok {
			if whiteouts[directory] {
				return true
			}
		}
		file = directory
	}
	return false
}

func checkWhitelist(target string) bool {
	for _, w := range constants.Whitelist {
		if strings.HasPrefix(target, w) {
			return true
		}
	}
	return false
}
