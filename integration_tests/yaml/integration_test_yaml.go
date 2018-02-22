package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

var tests = []struct {
	description    string
	dockerfilePath string
	context        string
	repo           string
}{
	{
		description:    "test extract filesystem",
		dockerfilePath: "dockerfiles/Dockerfile",
		context:        "dockerfiles/",
		repo:           "extract-filesystem",
	},
}

type step struct {
	Name string
	Args []string
	Env  []string
}

type testyaml struct {
	Steps []step
}

var executorImage = "gcr.io/kbuild-project/executor:latest"

func main() {

	// // Go install
	// installTests := step{
	// 	Name: "gcr.io/cloud-builders/go:debian",
	// 	Args: []string{"build", "integration.go"},
	// 	Env:  []string{"GOPATH=/", "PATH=/builder/bin:/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/workspace"},
	// }
	// First, copy container-diff in
	containerDiffStep := step{
		Name: "gcr.io/cloud-builders/gsutil",
		Args: []string{"cp", "gs://container-diff/latest/container-diff-linux-amd64", "."},
	}
	containerDiffPermissions := step{
		Name: "ubuntu",
		Args: []string{"chmod", "+x", "container-diff-linux-amd64"},
	}
	// Pull executor image
	pullExecutor := step{
		Name: "gcr.io/cloud-builders/docker",
		Args: []string{"pull", executorImage},
	}

	y := testyaml{
		Steps: []step{containerDiffStep, containerDiffPermissions, pullExecutor},
	}
	for _, test := range tests {
		// First, build the image with docker
		dockerBuild := step{
			Name: "gcr.io/cloud-builders/docker",
			Args: []string{"build", "-t", "gcr.io/kbuild-test/docker-" + test.repo, "-f", test.dockerfilePath, test.context},
		}

		// Then, buld the image with kbuild and commit it
		kbuild := step{
			Name: "gcr.io/cloud-builders/docker",
			Args: []string{"run", "-v", "/workspace/" + test.dockerfilePath + ":/dockerfile/Dockerfile", "--name", "test", executorImage, "/work-dir/executor"},
		}

		commit := step{
			Name: "gcr.io/cloud-builders/docker",
			Args: []string{"commit", "test", "gcr.io/kbuild-test/kbuild-" + test.repo},
		}

		// Then, push both images
		// pushDockerBuild := step{
		// 	Name: "gcr.io/cloud-builders/docker",
		// 	Args: []string{"push", "gcr.io/kbuild-test/docker-" + test.repo},
		// }
		// pushKbuild := step{
		// 	Name: "gcr.io/cloud-builders/docker",
		// 	Args: []string{"push", "gcr.io/kbuild-test/kbuild-" + test.repo},
		// }

		y.Steps = append(y.Steps, dockerBuild, kbuild, commit)
	}

	integrationTests := step{
		Name: "gcr.io/cloud-builders/go:debian",
		Args: []string{"test", "integration_test.go"},
		Env:  []string{"GOPATH=/", "PATH=/builder/bin:/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/workspace"},
	}
	y.Steps = append(y.Steps, integrationTests)

	d, _ := yaml.Marshal(&y)
	fmt.Println(string(d))
}
