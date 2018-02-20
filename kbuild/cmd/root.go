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
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"github.com/containers/image/docker"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var (
	dockerfilePath string
	name           string
	srcContext     string
	logLevel       string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&dockerfilePath, "dockerfile", "d", "Dockerfile", "Path to the dockerfile to be built.")
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
		if err := util.SetLogLevel(logLevel); err != nil {
			return err
		}
		return checkFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func checkFlags() error {
	if _, err := os.Stat(srcContext); err != nil {
		return errors.Wrap(err, "please provide valid source context with --context or -c flag")
	}
	if _, err := docker.ParseReference("//" + name); err != nil {
		return errors.Wrap(err, "please provide valid registry name for the final image with the --name or -n flag")
	}
	if _, err := os.Stat(dockerfilePath); err != nil {
		return errors.Wrap(err, "please provide valid path to Dockerfile with --dockerfile or -d flag")
	}
	return nil
}
