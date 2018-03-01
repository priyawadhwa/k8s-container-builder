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
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/contexts/dest"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type CopyCommand struct {
	cmd           *instructions.CopyCommand
	context       dest.Context
	snapshotFiles []string
}

func (c *CopyCommand) ExecuteCommand() error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	logrus.Infof("cmd: copy %s", srcs)
	logrus.Infof("dest: %s", dest)

	if err := c.checkContext(); err != nil {
		return err
	}

	if containsWildcards(srcs) {
		return c.executeWithWildcards()
	}
	// If there are multiple sources, the destination must be a directory
	if len(srcs) > 1 && !isDir(dest) {
		return errors.New("When specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
	}
	// If destination is not a directory, copy over the file into the destination
	if !isDir(dest) {
		src := filepath.Clean(srcs[0])
		files, err := c.context.GetFilesFromSource(src)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return errors.New("No source files specified for this command.")
		}
		if len(files) > 1 {
			return errors.New("When specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
		}
		for file, contents := range files {
			logrus.Infof("Copying from %s to %s", file, dest)
			if err := util.CreateFile(dest, contents); err != nil {
				return err
			}
			c.snapshotFiles = append(c.snapshotFiles, dest)
		}
		return nil
	}
	// Otherwise, go through each src, and copy over the files into dest
	for _, src := range srcs {
		src = filepath.Clean(src)
		files, err := c.context.GetFilesFromSource(src)
		if err != nil {
			return err
		}
		for file, contents := range files {
			relPath, err := filepath.Rel(src, file)
			if err != nil {
				return err
			}
			if relPath == "." {
				relPath = filepath.Base(file)
			}
			destPath := filepath.Join(dest, relPath)
			logrus.Infof("Copying from %s to %s", file, destPath)
			err = util.CreateFile(destPath, contents)
			if err != nil {
				return err
			}
			c.snapshotFiles = append(c.snapshotFiles, destPath)
		}
	}
	return nil
}

func (c *CopyCommand) executeWithWildcards() error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	// Get all files from the source, since each needs to be matched against wildcards
	files, err := c.context.GetFilesFromSource("")
	if err != nil {
		return err
	}
	matchedFiles, err := getMatchedFiles(srcs, files)
	logrus.Info(matchedFiles)
	if err != nil {
		return err
	}
	if !isDir(dest) {
		// If destination is not a directory, make sure only 1 file was matched
		totalFiles := 0
		for _, srcFiles := range matchedFiles {
			totalFiles += len(srcFiles)
		}
		if totalFiles == 0 {
			return errors.New("No source files specified for this command.")
		}
		if totalFiles > 1 {
			return errors.New("When specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
		}
		// Then, copy over the file to the destination
		for _, srcFiles := range matchedFiles {
			for _, file := range srcFiles {
				logrus.Infof("Copying %s into file %s", file, dest)
				if err := util.CreateFile(dest, files[file]); err != nil {
					return err
				}
				c.snapshotFiles = append(c.snapshotFiles, dest)
			}
		}
	}
	// Otherwise, destination is a directory, and we copy over all matched files
	for _, srcFiles := range matchedFiles {
		for _, file := range srcFiles {
			// Join destination and filename to create final path for the file
			destPath := filepath.Join(dest, filepath.Base(file))
			err = util.CreateFile(destPath, files[file])
			if err != nil {
				return err
			}
			c.snapshotFiles = append(c.snapshotFiles, destPath)
		}
	}
	return nil
}

func (c *CopyCommand) GetSnapshotFiles() []string {
	return c.snapshotFiles
}

func (c *CopyCommand) checkContext() error {
	if c.cmd.From != "" {
		filepath, err := util.GetImageTar(c.cmd.From)
		if err != nil {
			return err
		}
		logrus.Debugf("Using source context %s", filepath)
		c.context = dest.GetTarContext(filepath)
	}
	return nil
}

func isDir(path string) bool {
	return strings.HasSuffix(path, "/")
}

func containsWildcards(paths []string) bool {
	for _, path := range paths {
		for i := 0; i < len(path); i++ {
			ch := path[i]
			// These are the wildcards that correspond to filepath.Match
			if ch == '*' || ch == '?' || ch == '[' {
				return true
			}
		}
	}
	return false
}

func getMatchedFiles(srcs []string, files map[string][]byte) (map[string][]string, error) {
	f := make(map[string][]string)
	for _, src := range srcs {
		src = filepath.Clean(src)
		matchedFiles := []string{}
		for file := range files {
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
