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
	"fmt"
	"gopkg.in/yaml.v2"
)

const (
	executorImage           = "executor-image"
	executorCommand         = "/kbuild/executor"
	dockerImage             = "gcr.io/cloud-builders/docker"
	ubuntuImage             = "ubuntu"
	testRepo                = "gcr.io/kbuild-test/"
	dockerPrefix            = "docker-"
	kbuildPrefix            = "kbuild-"
	daemonPrefix            = "daemon://"
	containerDiffOutputFile = "container-diff.json"
	kbuildTestBucket        = "kbuild-test-bucket"
	buildcontextPath        = "/workspace/integration_tests"
	dockerfilesPath         = "/workspace/integration_tests/dockerfiles"
)

var fileTests = []struct {
	description    string
	dockerfilePath string
	configPath     string
	dockerContext  string
	kbuildContext  string
	repo           string
}{
	{
		description:    "test extract filesystem",
		dockerfilePath: "/workspace/integration_tests/dockerfiles/Dockerfile_test_extract_fs",
		configPath:     "/workspace/integration_tests/dockerfiles/config_test_extract_fs.json",
		dockerContext:  dockerfilesPath,
		kbuildContext:  dockerfilesPath,
		repo:           "extract-filesystem",
	},
	{
		description:    "test run",
		dockerfilePath: "/workspace/integration_tests/dockerfiles/Dockerfile_test_run",
		configPath:     "/workspace/integration_tests/dockerfiles/config_test_run.json",
		dockerContext:  dockerfilesPath,
		kbuildContext:  dockerfilesPath,
		repo:           "test-run",
	},
	{
		description:    "test run no files changed",
		dockerfilePath: "/workspace/integration_tests/dockerfiles/Dockerfile_test_run_2",
		configPath:     "/workspace/integration_tests/dockerfiles/config_test_run_2.json",
		dockerContext:  dockerfilesPath,
		kbuildContext:  dockerfilesPath,
		repo:           "test-run-2",
	},
	{
		description:    "test copy",
		dockerfilePath: "/workspace/integration_tests/dockerfiles/Dockerfile_test_copy",
		configPath:     "/workspace/integration_tests/dockerfiles/config_test_copy.json",
		dockerContext:  buildcontextPath,
		kbuildContext:  buildcontextPath,
		repo:           "test-copy",
	},
	{
		description:    "test bucket build context",
		dockerfilePath: "/workspace/integration_tests/dockerfiles/Dockerfile_test_copy",
		configPath:     "/workspace/integration_tests/dockerfiles/config_test_bucket_buildcontext.json",
		dockerContext:  buildcontextPath,
		kbuildContext:  kbuildTestBucket,
		repo:           "test-bucket-buildcontext",
	},
}

var structureTests = []struct {
	description           string
	dockerfilePath        string
	structureTestYamlPath string
	dockerBuildContext    string
	kbuildContext         string
	repo                  string
}{
	{
		description:           "test env",
		dockerfilePath:        "/workspace/integration_tests/dockerfiles/Dockerfile_test_env",
		repo:                  "test-env",
		dockerBuildContext:    dockerfilesPath,
		kbuildContext:         dockerfilesPath,
		structureTestYamlPath: "/workspace/integration_tests/dockerfiles/test_env.yaml",
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

func main() {

	// First, copy container-diff in
	containerDiffStep := step{
		Name: "gcr.io/cloud-builders/gsutil",
		Args: []string{"cp", "gs://container-diff/latest/container-diff-linux-amd64", "."},
	}
	containerDiffPermissions := step{
		Name: ubuntuImage,
		Args: []string{"chmod", "+x", "container-diff-linux-amd64"},
	}
	structureTestsStep := step{
		Name: "gcr.io/cloud-builders/gsutil",
		Args: []string{"cp", "gs://container-structure-test/latest/container-structure-test", "."},
	}
	structureTestPermissions := step{
		Name: ubuntuImage,
		Args: []string{"chmod", "+x", "container-structure-test"},
	}

	GCSBucketTarBuildContext := step{
		Name: ubuntuImage,
		Args: []string{"tar", "-C", "/workspace/integration_tests/", "-cf", "/workspace/kbuild.tar", "."},
	}
	uploadTarBuildContext := step{
		Name: "gcr.io/cloud-builders/gsutil",
		Args: []string{"cp", "/workspace/kbuild.tar", "gs://kbuild-test-bucket/"},
	}

	// Build executor image
	buildExecutorImage := step{
		Name: dockerImage,
		Args: []string{"build", "-t", executorImage, "-f", "integration_tests/executor/Dockerfile", "."},
	}
	y := testyaml{
		Steps: []step{containerDiffStep, containerDiffPermissions, structureTestsStep, structureTestPermissions, GCSBucketTarBuildContext, uploadTarBuildContext, buildExecutorImage},
	}
	for _, test := range fileTests {
		// First, build the image with docker
		dockerImageTag := testRepo + dockerPrefix + test.repo
		dockerBuild := step{
			Name: dockerImage,
			Args: []string{"build", "-t", dockerImageTag, "-f", test.dockerfilePath, test.dockerContext},
		}

		// Then, buld the image with kbuild
		kbuildImage := testRepo + kbuildPrefix + test.repo
		kbuild := step{
			Name: executorImage,
			Args: []string{executorCommand, "--destination", kbuildImage, "--dockerfile", test.dockerfilePath, "--context", test.kbuildContext},
		}

		// Pull the kbuild image
		pullKbuildImage := step{
			Name: dockerImage,
			Args: []string{"pull", kbuildImage},
		}

		daemonDockerImage := daemonPrefix + dockerImageTag
		daemonKbuildImage := daemonPrefix + kbuildImage
		// Run container diff on the images
		args := "container-diff-linux-amd64 diff " + daemonDockerImage + " " + daemonKbuildImage + " --type=file -j >" + containerDiffOutputFile
		containerDiff := step{
			Name: ubuntuImage,
			Args: []string{"sh", "-c", args},
			Env:  []string{"PATH=/workspace:/bin"},
		}

		catContainerDiffOutput := step{
			Name: ubuntuImage,
			Args: []string{"cat", containerDiffOutputFile},
		}
		compareOutputs := step{
			Name: ubuntuImage,
			Args: []string{"cmp", test.configPath, containerDiffOutputFile},
		}

		y.Steps = append(y.Steps, dockerBuild, kbuild, pullKbuildImage, containerDiff, catContainerDiffOutput, compareOutputs)
	}

	for _, test := range structureTests {

		// First, build the image with docker
		dockerImageTag := testRepo + dockerPrefix + test.repo
		dockerBuild := step{
			Name: dockerImage,
			Args: []string{"build", "-t", dockerImageTag, "-f", test.dockerfilePath, test.dockerBuildContext},
		}

		// Build the image with kbuild
		kbuildImage := testRepo + kbuildPrefix + test.repo
		kbuild := step{
			Name: executorImage,
			Args: []string{executorCommand, "--destination", kbuildImage, "--dockerfile", test.dockerfilePath, "--context", test.kbuildContext},
		}
		// Pull the kbuild image
		pullKbuildImage := step{
			Name: dockerImage,
			Args: []string{"pull", kbuildImage},
		}
		// Run structure tests on the kbuild and docker image
		args := "container-structure-test -image " + kbuildImage + " " + test.structureTestYamlPath
		structureTest := step{
			Name: ubuntuImage,
			Args: []string{"sh", "-c", args},
			Env:  []string{"PATH=/workspace:/bin"},
		}
		args = "container-structure-test -image " + dockerImageTag + " " + test.structureTestYamlPath
		dockerStructureTest := step{
			Name: ubuntuImage,
			Args: []string{"sh", "-c", args},
			Env:  []string{"PATH=/workspace:/bin"},
		}

		y.Steps = append(y.Steps, dockerBuild, kbuild, pullKbuildImage, structureTest, dockerStructureTest)
	}

	d, _ := yaml.Marshal(&y)
	fmt.Println(string(d))
}
