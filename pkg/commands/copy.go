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
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/containers/image/manifest"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type CopyCommand struct {
	cmd           *instructions.CopyCommand
	buildcontext  string
	snapshotFiles []string
}

func (c *CopyCommand) ExecuteCommand(config *manifest.Schema2Config) error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	logrus.Infof("cmd: copy %s", srcs)
	logrus.Infof("dest: %s", dest)

	srcs, err := util.ResolveSources(srcs, c.buildcontext, config.WorkingDir)
	if err != nil {
		return err
	}
	srcMap, err := util.SourcesToFilesMap(srcs, c.buildcontext)
	if err != nil {
		return err
	}
	if err := util.IsSrcsValid(c.cmd.SourcesAndDest, srcMap); err != nil {
		return err
	}
	for src, files := range srcMap {
		for _, file := range files {
			fi, err := os.Stat(filepath.Join(c.buildcontext, file))
			if err != nil {
				return err
			}
			destPath, err := util.RelativeFilepath(file, src, dest, config.WorkingDir, c.buildcontext)
			if err != nil {
				return err
			}
			if fi.IsDir() {
				logrus.Infof("Creating directory %s", destPath)
				if err := os.MkdirAll(destPath, fi.Mode()); err != nil {
					return err
				}
			} else {
				logrus.Infof("Copying file %s to %s", file, destPath)
				contents, err := ioutil.ReadFile(filepath.Join(c.buildcontext, file))
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

// FilesToSnapshot should return an empty array if still nil; no files were changed
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
