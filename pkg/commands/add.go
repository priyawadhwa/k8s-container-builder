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
	"os"
	"path/filepath"
	"strings"
)

type AddCommand struct {
	cmd           *instructions.AddCommand
	buildcontext  string
	snapshotFiles []string
}

// ExecuteCommand executes the ADD command
// Special stuff about ADD:
// 	1. If <src> is a remote file URL:
// 		- destination will have permissions of 0600
// 		- If remote file has HTTP Last-Modified header, we set the mtime of the file to that timestamp
// 		- mtime should not be included in determining whether the file has been changed
// 		- If dest doesn't end with a slash, the filepath is inferred to be <dest>/<filename>
// 	2. If <src> is a local tar archive:
// 		-If <src> is a local tar archive, it is unpacked at the dest with, like 'tar -x' would
func (a *AddCommand) ExecuteCommand(config *manifest.Schema2Config) error {
	srcs := a.cmd.SourcesAndDest[:len(a.cmd.SourcesAndDest)-1]
	dest := a.cmd.SourcesAndDest[len(a.cmd.SourcesAndDest)-1]

	logrus.Infof("cmd: Add %s", srcs)
	logrus.Infof("dest: %s", dest)

	// First, resolve any environment replacement
	resolvedEnvs, err := util.ResolveEnvironmentReplacementList(a.AddToString(), a.cmd.SourcesAndDest, config.Env, true)
	if err != nil {
		return err
	}
	dest = resolvedEnvs[len(a.cmd.SourcesAndDest)-1]
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(config.WorkingDir, dest)
	}
	// Get a map of [src]:[files rooted at src]
	srcMap, err := util.ResolveSources(resolvedEnvs, a.buildcontext)
	if err != nil {
		return err
	}
	// If any of the sources are local tar archives:
	// 	1. Unpack them to the specified destination
	// 	2. Remove them as sources that need to be copied over

	for _, files := range srcMap {
		for _, file := range files {
			// If file is a local tar archive, then we unpack it to dest
			filePath := filepath.Join(a.buildcontext, file)
			isFilenameSource, err := isFilenameSource(srcMap, file)
			if err != nil {
				return err
			}
			if isFilenameSource && util.IsFileLocalTarArchive(filePath) {
				logrus.Infof("Unpacking local tar archive %s to %s", file, dest)
				if err := util.UnpackLocalTarArchive(filePath, dest); err != nil {
					return err
				}
				// Add the unpacked files to the snapshotter
				filesAdded, err := util.Files(dest)
				if err != nil {
					return err
				}
				logrus.Debugf("Added %v from local tar archive %s", filesAdded, file)
				a.snapshotFiles = append(a.snapshotFiles, filesAdded...)
			}
		}
	}

	// If any of the sources is a remote file URL:
	// 	1. Copy over the file to the specified destination
	// 	2. Remove as a source that needs to be copied over

	// With the remaining "normal" sources, create and execute a standard copy command

	// For each source, iterate through each file within and Add it over
	for src, files := range srcMap {
		for _, file := range files {
			filePath := filepath.Join(a.buildcontext, file)
			fi, err := os.Stat(filePath)
			if err != nil {
				return err
			}
			// If file is a local tar archive, then we unpack it to dest
			isFilenameSource, err := isFilenameSource(srcMap, file)
			if err != nil {
				return err
			}
			if isFilenameSource && util.IsFileLocalTarArchive(filePath) {
				logrus.Infof("Unpacking local tar archive %s to %s", file, dest)
				if !filepath.IsAbs(dest) {
					dest = filepath.Join(config.WorkingDir, dest)
				}
				if err := util.UnpackLocalTarArchive(filePath, dest); err != nil {
					return err
				}
				// Add the unpacked files to the snapshotter
				filesAdded, err := util.Files(dest)
				if err != nil {
					return err
				}
				logrus.Infof("Added %v from local tar archive %s", filesAdded, file)
				a.snapshotFiles = append(a.snapshotFiles, filesAdded...)
				continue
			}
			destPath, err := util.DestinationFilepath(file, src, dest, config.WorkingDir, a.buildcontext)
			if err != nil {
				return err
			}
			// If source file is a directory, we want to create a directory ...
			if fi.IsDir() {
				logrus.Infof("Creating directory %s", destPath)
				if err := os.MkdirAll(destPath, fi.Mode()); err != nil {
					return err
				}
			} else {
				// ... Else, we want to Add over a file
				logrus.Infof("Adding file %s to %s", file, destPath)
				srcFile, err := os.Open(filepath.Join(a.buildcontext, file))
				if err != nil {
					return err
				}
				defer srcFile.Close()
				if err := util.CreateFile(destPath, srcFile, fi.Mode()); err != nil {
					return err
				}
			}
			// Append the destination file to the list of files that should be snapshotted later
			a.snapshotFiles = append(a.snapshotFiles, destPath)
		}
	}
	return nil
}

func isFilenameSource(srcMap map[string][]string, fileName string) (bool, error) {
	for src := range srcMap {
		matched, err := filepath.Match(src, fileName)
		if err != nil {
			return false, err
		}
		if matched || (src == fileName) {
			return true, nil
		}
	}
	return false, nil
}

// AddToString returns a string of the command
func (a *AddCommand) AddToString() string {
	Add := []string{"ADD"}
	return strings.Join(append(Add, a.cmd.SourcesAndDest...), " ")
}

// FilesToSnapshot should return an empty array if still nil; no files were changed
func (a *AddCommand) FilesToSnapshot() []string {
	return a.snapshotFiles
}

// CreatedBy returns some information about the command for the image config
func (a *AddCommand) CreatedBy() string {
	return strings.Join(a.cmd.SourcesAndDest, " ")
}
