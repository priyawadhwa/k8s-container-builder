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
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/image"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/sirupsen/logrus"
	"os"
)

// EnvCommand struct for Docker ENV command
type EnvCommand struct {
	cmd *instructions.EnvCommand
}

// ExecuteCommand sets the env variables
func (e *EnvCommand) ExecuteCommand() error {
	logrus.Info("cmd: ENV")
	newEnvs := e.cmd.Env
	for _, pair := range newEnvs {
		fmt.Printf("Setting environment variable %s:%s", pair.Key, pair.Value)
		if err := os.Setenv(pair.Key, pair.Value); err != nil {
			return err
		}
	}
	return e.addEnvToConfig()
}

func (e *EnvCommand) addEnvToConfig() error {
	newEnvs := e.cmd.Env
	envs := image.Env()
	for _, pair := range newEnvs {
		envs[pair.Key] = pair.Value
	}
	image.SetEnv(envs)
	return nil
}

func (e *EnvCommand) GetSnapshotFiles() []string {
	return []string{}
}
