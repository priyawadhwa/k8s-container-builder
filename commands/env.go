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
	"github.com/GoogleCloudPlatform/k8s-container-builder/appender"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"os"
)

// EnvCommand struct for Docker ENV command
type EnvCommand struct {
	cmd *instructions.EnvCommand
}

// ExecuteCommand sets the env variables
func (e EnvCommand) ExecuteCommand() error {
	fmt.Println("cmd: ENV")
	envVars := e.cmd.Env
	for _, pair := range envVars {
		fmt.Printf("Setting environment variable %s:%s", pair.Key, pair.Value)
		if err := os.Setenv(pair.Key, pair.Value); err != nil {
			return err
		}
		if err := appender.MutableSource.AddEnv(pair.Key, pair.Value); err != nil {
			return err
		}
	}
	return nil
}
