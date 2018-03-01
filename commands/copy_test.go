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
	"github.com/pkg/errors"
	"sort"
	"testing"
)

var testCases = []struct {
	name   string
	srcs   []string
	files  map[string][]byte
	output map[string][]string
}{
	{
		name: "multiple sources",
		srcs: []string{
			"pkg/*",
			"dir/a*.go",
		},
		files: map[string][]byte{
			"pkg/a":         nil,
			"pkg/b/c":       nil,
			"test":          nil,
			"dir/apple.go":  nil,
			"dir/banana.go": nil,
		},
		output: map[string][]string{
			"pkg/*": {
				"pkg/a",
			},
			"dir/a*.go": {
				"dir/apple.go",
			},
		},
	},
	{
		name: "wildcard and normal srcs",
		srcs: []string{
			"pkg/*",
			"dir/a*.go",
			"pkg/",
		},
		files: map[string][]byte{
			"pkg/a":         nil,
			"pkg/b/c":       nil,
			"test":          nil,
			"dir/apple.go":  nil,
			"dir/banana.go": nil,
		},
		output: map[string][]string{
			"pkg/*": {
				"pkg/a",
			},
			"dir/a*.go": {
				"dir/apple.go",
			},
			"pkg": {
				"pkg/a",
				"pkg/b/c",
			},
		},
	},
	{
		name: "no match",
		srcs: []string{
			"pkg/*",
			"test/",
		},
		files: map[string][]byte{
			"pkg/a": nil,
		},
		output: map[string][]string{
			"pkg/*": {
				"pkg/a",
			},
			"test": {},
		},
	},
	{
		name: "one file",
		srcs: []string{
			"pkg/*",
		},
		files: map[string][]byte{
			"/pkg/a": nil,
		},
		output: map[string][]string{
			"pkg/*": {
				"/pkg/a",
			},
		},
	},
}

func TestCopy_getMatchedFiles(t *testing.T) {
	for _, tc := range testCases {
		output, err := getMatchedFiles(tc.srcs, tc.files)
		for _, value := range output {
			sort.Strings(value)
		}
		testutil.CheckErrorAndDeepEqual(t, false, errors.Wrap(err, tc.name), tc.output, output)
	}
}
