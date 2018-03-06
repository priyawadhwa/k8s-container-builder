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

package cmd

import (
	"github.com/GoogleCloudPlatform/k8s-container-builder/commands"
	"github.com/GoogleCloudPlatform/k8s-container-builder/contexts/dest"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/dockerfile"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/env"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/image"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/snapshot"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	dockerfilePath string
	name           string
	srcContext     string
	logLevel       string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&dockerfilePath, "dockerfile", "d", "/workspace/Dockerfile", "Path to the dockerfile to be built.")
	RootCmd.PersistentFlags().StringVarP(&srcContext, "context", "c", "", "Path to the dockerfile build context.")
	RootCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Registry the final image should be pushed to (ex: gcr.io/test/example:latest)")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "verbosity", "v", constants.DefaultLogLevel, "Log level (debug, info, warn, error, fatal, panic")
}

var RootCmd = &cobra.Command{
	Use: "executor",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return util.SetLogLevel(logLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := execute(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
	},
}

func execute() error {
	// Initialize whitelist
	if err := util.InitializeWhitelist(); err != nil {
		return err
	}

	// Read and parse dockerfile
	b, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return err
	}

	stages, err := dockerfile.Parse(b)
	if err != nil {
		return err
	}

	for index, stage := range stages {

		baseImage := stage.BaseName

		finalStage := (index + 1) == len(stages)
		if finalStage {
			// Initialize source image
			logrus.Info("Initializing source image")
			if err := image.InitializeSourceImage(baseImage); err != nil {
				logrus.Fatalf("Unable to intitalize source images %s: %v", baseImage, err)
			}
		}
		logrus.Infof("Extracting filesystem for %s...", baseImage)
		err = util.ExtractFileSystemFromImage(baseImage)
		if err != nil {
			return err
		}
		l := snapshot.NewLayeredMap(util.Hasher())
		snapshotter := snapshot.NewSnapshotter(l, constants.RootDir)

		// Take initial snapshot
		if err := snapshotter.Init(); err != nil {
			return err
		}

		// Get context information
		context := dest.GetContext(srcContext)

		for _, cmd := range stage.Commands {
			dockerCommand := commands.GetCommand(cmd, context)
			if err := dockerCommand.ExecuteCommand(); err != nil {
				return err
			}
			var contents []byte
			if dockerCommand.GetSnapshotFiles() != nil {
				logrus.Info("Taking snapshot of specific files now.")
				contents, err = snapshotter.TakeSnapshotOfFiles(dockerCommand.GetSnapshotFiles())
				if err != nil {
					return err
				}
				if contents == nil {
					logrus.Info("Contents are empty, continue.")
					continue
				}
			} else {
				logrus.Info("Taking generic snapshot now.")
				c, filesAdded, err := snapshotter.TakeSnapshot()
				contents = c
				if err != nil {
					return err
				}
				if !filesAdded {
					logrus.Info("No files were changed in this command, appending empty layer to config history.")
					image.AppendEmptyLayerToConfigHistory("kbuild")
					continue
				}
			}
			if finalStage {
				logrus.Info("Appending to source image")
				if err := image.AppendLayer(contents); err != nil {
					return err
				}
			}
		}
		if finalStage {
			// Save environment variables
			env.SetEnvironmentVariables(baseImage)
			continue
		}
		// Now package up filesystem as tarball
		tarballFiles, err := dockerfile.GetMultiStageDependencies(index, stage.Name, stages)
		logrus.Infof("Saving these files from stage %v: %s", index, tarballFiles)
		if err != nil {
			return err
		}
		if err := util.SaveFilesToTarball(stage.Name, index, tarballFiles); err != nil {
			return err
		}
		// Then, delete filesystem
		util.DeleteFileSystem()
	}
	// Push the image
	return image.PushImage(name)
}
