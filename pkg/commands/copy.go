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

type Source struct {
	src   string
	files []string
}

func (c *CopyCommand) ExecuteCommand(config *manifest.Schema2Config) error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	logrus.Infof("cmd: copy %s", srcs)
	logrus.Infof("dest: %s", dest)

	// Current working directory
	// TODO: Update for WORKDIR command
	cwd := constants.RootDir

	srcMap, err := c.MatchedFiles()
	logrus.Info(srcMap)
	if err != nil {
		return err
	}

	if err := c.checkSources(srcMap, dest); err != nil {
		return err
	}

	for src, files := range srcMap {
		for _, file := range files {
			fi, err := c.buildcontext.Stat(file)
			if err != nil {
				return err
			}
			destPath := util.WildcardRelativeFilepath(file, dest)
			if !util.ContainsWildcards(srcs) {
				srcfi, err := c.buildcontext.Stat(src)
				if err != nil {
					return err
				}
				destPath, err = util.RelativeFilepath(file, src, cwd, dest, srcfi, fi)
				if err != nil {
					return err
				}
			}
			if fi.IsDir() {
				logrus.Infof("Creating directory %s", destPath)
				if err := os.MkdirAll(destPath, fi.Mode()); err != nil {
					return err
				}
			} else {
				logrus.Infof("Copying file %s to %s", file, destPath)
				contents, err := c.buildcontext.Contents(file)
				if err != nil {
					return err
				}
				if err := util.CreateFile(destPath, contents, fi.Mode()); err != nil {
					return err
				}
			}
			c.snapshotFiles = append(c.snapshotFiles, destPath)
		}
	}
	return nil
}

// MatchedFiles returns a map of [src]:[files matching source]
func (c *CopyCommand) MatchedFiles() (map[string][]string, error) {
	srcMap := make(map[string][]string)
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	if util.ContainsWildcards(srcs) {
		logrus.Info("contains wildcards")
		files, err := c.buildcontext.Files("")
		if err != nil {
			return nil, err
		}
		return util.GetMatchedFiles(srcs, files)
	}
	for _, src := range srcs {
		src = filepath.Clean(src)
		files, err := c.buildcontext.Files(src)
		if err != nil {
			return nil, err
		}
		srcMap[src] = files
	}
	return srcMap, nil
}

// FilesToSnapshot returns nil for this command because we don't know which files
// have changed, so we snapshot the entire system.
func (c *CopyCommand) FilesToSnapshot() []string {
	if c.snapshotFiles == nil {
		return []string{}
	}
	return c.snapshotFiles
}

// Author returns some information about the command for the image config
func (c *CopyCommand) CreatedBy() string {
	return strings.Join(c.cmd.SourcesAndDest, " ")
}

func (c *CopyCommand) checkSources(srcMap map[string][]string, dest string) error {
	// If destination is a directory, return nil
	if util.IsDestDir(dest) {
		return nil
	}
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	wildcard := util.ContainsWildcards(srcs)
	// If no wildcards and multiple sources, return error
	if !wildcard {
		if len(srcs) > 1 {
			return errors.New("when specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
		}
		return nil
	}
	// If no wildcards, and source is dir, return niil
	totalFiles := 0
	for _, files := range srcMap {
		totalFiles += len(files)
	}
	if totalFiles == 0 {
		return errors.New("copy failed: no source files specified")
	}
	if totalFiles > 1 {
		return errors.New("when specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
	}
	return nil
}
