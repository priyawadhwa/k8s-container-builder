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
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

type RunCommand struct {
	cmd *instructions.RunCommand
}

func (r RunCommand) ExecuteCommand() error {
	var newCommand []string
	if r.cmd.PrependShell {
		newCommand = []string{"sh", "-c"}
		newCommand = append(newCommand, strings.Join(r.cmd.CmdLine, " "))
	} else {
		newCommand = r.cmd.CmdLine
	}
	return execute(newCommand)
}

func execute(c []string) error {
	logrus.Infof("cmd: ", c[0])
	logrus.Infof("args: ", c[1:])
	cmd := exec.Command(c[0], c[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Error: %s", output)
		return err
	}
	logrus.Infof("Output from %s %s\n", cmd.Path, cmd.Args)
	return nil
}
