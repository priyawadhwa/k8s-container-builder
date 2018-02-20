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
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	dockerfile string
	name       string
	srcContext string
	logLevel   string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&dockerfile, "dockerfile", "d", "Dockerfile", "Path to the dockerfile to be built.")
	RootCmd.PersistentFlags().StringVarP(&srcContext, "context", "c", "", "Path to the dockerfile build context.")
	RootCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Registry the final image should be pushed to (ex: gcr.io/test/example:latest)")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "verbosity", "v", constants.DefaultLogLevel, "Log level (debug, info, warn, error, fatal, panic")
}

var RootCmd = &cobra.Command{
	Use:   "kbuild",
	Short: "kbuild is a CLI tool for building container images with full Dockerfile support without the need for Docker",
	Long: `kbuild is a CLI tool for building container images with full Dockerfile support. It doesn't require Docker,
			and builds the images in a Kubernetes cluster before pushing the final image to a registry.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return errors.Wrap(err, "parsing log level")
		}
		logrus.SetLevel(lvl)
		if err := checkFlags(); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func checkFlags() error {
	if srcContext == "" {
		return errors.Errorf("Please provide source context with --context or -c flag")
	}
	if name == "" {
		return errors.Errorf("Please provide name of final image with --name or -n flag")
	}
	if _, err := os.Stat(dockerfile); err != nil {
		return errors.Wrap(err, "please provide valid path to Dockerfile with --dockerfile or -d flag")
	}
	return nil
}
