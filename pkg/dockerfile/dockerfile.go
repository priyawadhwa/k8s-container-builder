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

package dockerfile

import (
	"bytes"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"

	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/docker/docker/builder/dockerfile/parser"
)

// Parse parses the contents of a Dockerfile and returns a list of commands
func Parse(b []byte) ([]instructions.Stage, error) {
	p, err := parser.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	stages, _, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}
	return stages, err
}

// Return a list of dependencies in stage index used later on in the dockerfile
func GetMultiStageDependencies(index int, name string, stages []instructions.Stage) ([]string, error) {

	var files []string

	for stageIndex, stage := range stages {
		if stageIndex <= index {
			continue
		}
		for _, cmd := range stage.Commands {
			switch c := cmd.(type) {
			case *instructions.CopyCommand:
				logrus.Debug("copy")
				logrus.Debug(c.From)
				if c.From == strconv.Itoa(index) || c.From == name {
					if util.ContainsWildcards(c.Sources()) {
						//TODO : Fill this out
					} else {
						for _, src := range c.Sources() {
							// Get all files from the source
							f, err := util.Files(src)
							if err != nil {
								return nil, err
							}
							for _, file := range f {
								fi, err := os.Stat(file)
								if err != nil {
									return nil, err
								}
								if fi.IsDir() {
									continue
								}
								logrus.Infof("Appending %s", file)
								files = append(files, file)
							}
						}
					}
				}
			}
		}
	}
	return files, nil
}
