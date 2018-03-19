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
package constants

const (
	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// RootDir is the path to the root directory
	RootDir = "/"

	// WorkspaceDir is the path to the workspace
	WorkspaceDir = "/work-dir/"

	WhitelistPath = "/proc/self/mountinfo"

	Author = "kbuild"

	ConfigPath = "/root/.docker/config.json"
)

// DefaultEnvVariables are the default env variables set in the container
var DefaultEnvVariables = map[string]string{
	"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root:/work-dir",
	"HOME": "/root",
}
