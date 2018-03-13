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

package commands

import (
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/buildcontext"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/containers/image/manifest"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type CopyCommand struct {
	cmd           *instructions.CopyCommand
	buildcontext  buildcontext.BuildContext
	snapshotFiles []string
}

func (c *CopyCommand) ExecuteCommand(config *manifest.Schema2Config) error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	logrus.Infof("cmd: copy %s", srcs)
	logrus.Infof("dest: %s", dest)

	c.snapshotFiles = []string{}
	cwd := constants.RootDir

	if util.ContainsWildcards(srcs) {
		return nil
		// return c.executeWithWildcards()
	}

	// If there are multiple sources, the destination must be a directory
	if len(srcs) > 1 && !IsDir(dest) {
		return errors.New("when specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
	}
	// If destination is not a directory, copy over the file into the destination
	if !IsDir(dest) {
		return c.CopySingleFile(srcs[0], dest, cwd, srcs)
	}
	// Otherwise, go through each src, and copy over the files into dest
	for _, src := range srcs {
		src = filepath.Clean(src)
		files, err := c.buildcontext.Files(src)
		if err != nil {
			return err
		}
		for _, file := range files {
			if c.buildcontext.Exists(file) {
				destPath, err := RelativePath(src, file, cwd, dest)
				if err != nil {
					return err
				}
				fi, err := c.buildcontext.Stat(file)
				if err != nil {
					return err
				}
				if fi.IsDir() {
					if err := os.MkdirAll(destPath, fi.Mode()); err != nil {
						return err
					}
					logrus.Infof("Creating directory %s", destPath)
				} else {
					contents, err := c.buildcontext.Contents(file)
					if err != nil {
						return err
					}
					if err := util.CreateFile(destPath, contents, fi.Mode()); err != nil {
						return err
					}
					logrus.Infof("Copied files %s to %s", file, destPath)
				}
				c.snapshotFiles = append(c.snapshotFiles, destPath)
			}
		}
	}
	return nil
}

// func (c *CopyCommand) executeWithWildcards() error {
// 	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
// 	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

// 	if !IsDir(dest) {
// 		return c.CopySingleFile("", dest, srcs)
// 	}
// 	// Otherwise, destination is a directory, and we copy over all matched files
// 	// Get all files from the source, since each needs to be matched against wildcards
// 	files, err := c.buildcontext.GetFilesFromPath("")
// 	if err != nil {
// 		return err
// 	}
// 	matchedFiles, err := util.GetMatchedFiles(srcs, files)
// 	logrus.Info(matchedFiles)
// 	if err != nil {
// 		return err
// 	}
// 	for _, srcFiles := range matchedFiles {
// 		for _, file := range srcFiles {
// 			// Join destination and filename to create final path for the file
// 			destPath := filepath.Join(dest, filepath.Base(file))
// 			err = util.CreateFile(destPath, files[file])
// 			if err != nil {
// 				return err
// 			}
// 			c.snapshotFiles = append(c.snapshotFiles, destPath)
// 		}
// 	}
// 	return nil
// }

// FilesToSnapshot returns nil for this command because we don't know which files
// have changed, so we snapshot the entire system.
func (c *CopyCommand) FilesToSnapshot() []string {
	return c.snapshotFiles
}

// Author returns some information about the command for the image config
func (c *CopyCommand) CreatedBy() string {
	return strings.Join(c.cmd.SourcesAndDest, " ")
}

func IsDir(path string) bool {
	return strings.HasSuffix(path, "/")
}

func (c *CopyCommand) CopySingleFile(path, dest, cwd string, srcs []string) error {
	path = filepath.Clean(path)
	files, err := c.buildcontext.Files(path)
	if err != nil {
		return err
	}
	matchedFiles, err := util.GetMatchedFiles(srcs, files)
	if err != nil {
		return err
	}

	totalFiles := 0
	for _, srcFiles := range matchedFiles {
		totalFiles += len(srcFiles)
	}
	if totalFiles == 0 {
		return errors.New("no source files specified for this command")
	}
	if totalFiles > 1 {
		return errors.New("when specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
	}
	// Then, copy over the file to the destination
	for _, srcFiles := range matchedFiles {
		for _, file := range srcFiles {
			if c.buildcontext.Exists(file) {
				fi, err := c.buildcontext.Stat(file)
				if err != nil {
					return err
				}
				dest = filepath.Join(cwd, dest)
				contents, err := c.buildcontext.Contents(file)
				if err != nil {
					return err
				}
				if err := util.CreateFile(dest, contents, fi.Mode()); err != nil {
					return err
				}
				logrus.Infof("Copied %s to %s", file, dest)
				c.snapshotFiles = append(c.snapshotFiles, dest)
			}
		}
	}
	return nil
}

func RelativePath(src, file, cwd, dest string) (string, error) {
	relPath, err := filepath.Rel(src, file)
	if err != nil {
		return "", err
	}
	if relPath == "." {
		relPath = filepath.Base(file)
	}
	destPath := filepath.Join(cwd, dest, relPath)
	return destPath, nil
}
