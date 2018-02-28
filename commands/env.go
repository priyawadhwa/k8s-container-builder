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
	}
	return e.addEnvToConfig()
}

func (e EnvCommand) addEnvToConfig() error {
	currentEnvs := 
}

func (m *MutableSource) AddEnv(key, value string) error {
	currentEnv := m.cfg.Schema2V1Image.Config.Env
	// First split into map of key:value pairs
	envMap := make(map[string]string)
	for _, e := range currentEnv {
		arr := strings.Split(e, "=")
		envMap[arr[0]] = arr[1]
	}
	envMap[key] = value
	var newEnv []string
	for key, value := range envMap {
		newEnv = append(newEnv, key+"="+value)
	}
	m.cfg.Schema2V1Image.Config.Env = newEnv
	return nil
}
