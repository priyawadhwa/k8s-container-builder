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
	"github.com/GoogleCloudPlatform/k8s-container-builder/testutil"
	"github.com/containers/image/manifest"
	"testing"
)

func TestUpdateExposedPorts(t *testing.T) {
	cfg := &manifest.Schema2Config{
		ExposedPorts: manifest.Schema2PortSet{
			"8080/tcp": {},
		},
	}

	ports := []string{
		"8080",
		"8081/tcp",
		"8082",
		"8083/udp",
	}

	expectedPorts := manifest.Schema2PortSet{
		"8080/tcp": {},
		"8081/tcp": {},
		"8082/tcp": {},
		"8083/udp": {},
	}

	err := updateExposedPorts(ports, cfg)
	testutil.CheckErrorAndDeepEqual(t, false, err, expectedPorts, cfg.ExposedPorts)
}

func TestInvalidProtocol(t *testing.T) {
	cfg := &manifest.Schema2Config{
		ExposedPorts: manifest.Schema2PortSet{},
	}

	ports := []string{
		"80/garbage",
	}

	err := updateExposedPorts(ports, cfg)
	testutil.CheckErrorAndDeepEqual(t, true, err, nil, nil)
}
