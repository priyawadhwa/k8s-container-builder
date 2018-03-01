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

package main

import (
	"flag"
	"github.com/GoogleCloudPlatform/k8s-container-builder/commands"
	"github.com/GoogleCloudPlatform/k8s-container-builder/contexts/dest"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/dockerfile"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/env"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/image"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/snapshot"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var dockerfilePath = flag.String("dockerfile", "/dockerfile/Dockerfile", "Path to Dockerfile.")
var source = flag.String("source", "", "Source context location")
var destImg = flag.String("dest", "", "Destination of final image")
var v = flag.String("verbosity", "info", "Logging verbosity")

func main() {
	flag.Parse()
	if err := setLogLevel(); err != nil {
		logrus.Fatal(err)
	}
	// Initialize whitelist
	if err := util.InitializeWhitelist(); err != nil {
		logrus.Fatal(err)
	}

	// Read and parse dockerfile
	b, err := ioutil.ReadFile(*dockerfilePath)
	if err != nil {
		logrus.Fatal(err)
	}

	stages, err := dockerfile.Parse(b)

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
			logrus.Fatal(err)
		}
		l := snapshot.NewLayeredMap(util.Hasher())
		snapshotter := snapshot.NewSnapshotter(l, constants.RootDir)

		// Take initial snapshot
		if err := snapshotter.Init(); err != nil {
			logrus.Fatal(err)
		}

		// Get context information
		context := dest.GetContext(*source)

		for _, cmd := range stage.Commands {
			dockerCommand := commands.GetCommand(cmd, context)
			if err := dockerCommand.ExecuteCommand(); err != nil {
				logrus.Fatal(err)
			}
			var contents []byte
			if dockerCommand.GetSnapshotFiles() != nil {
				logrus.Info("Taking snapshot of specific files now.")
				contents, err = snapshotter.TakeSnapshotOfFiles(dockerCommand.GetSnapshotFiles())
				if err != nil {
					logrus.Fatal(err)
				}
			} else {
				logrus.Info("Taking generic snapshot now.")
				contents, err = snapshotter.TakeSnapshot()
				if err != nil {
					logrus.Fatal(err)
				}
			}
			if contents == nil {
				logrus.Info("Contents are nil, continue.")
				continue
			}
			if finalStage {
				logrus.Info("Appending to source image")
				if err := image.AppendLayer(contents); err != nil {
					logrus.Fatal(err)
				}
			}
		}
		if finalStage {
			// Save environment variables
			env.SetEnvironmentVariables(baseImage)
			continue
		}
		// Now package up filesystem as tarball
		if err := util.SaveFileSystemAsTarball(stage.Name, index); err != nil {
			logrus.Fatal(err)
		}
		// Then, delete filesystem
		util.DeleteFileSystem()
	}

	// Push the image
	if err := image.PushImage(*destImg); err != nil {
		logrus.Fatal(err)
	}
	return
}

func setLogLevel() error {
	lvl, err := logrus.ParseLevel(*v)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)
	return nil
}
