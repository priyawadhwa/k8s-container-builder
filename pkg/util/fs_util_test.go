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

package util

import (
	"github.com/GoogleCloudPlatform/k8s-container-builder/testutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func Test_fileSystemWhitelist(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Error creating tempdir: %s", err)
	}
	fileContents := `
	228 122 0:90 / / rw,relatime - aufs none rw,si=f8e2406af90782bc,dio,dirperm1
	229 228 0:98 / /proc rw,nosuid,nodev,noexec,relatime - proc proc rw
	230 228 0:99 / /dev rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755
	231 230 0:100 / /dev/pts rw,nosuid,noexec,relatime - devpts devpts rw,gid=5,mode=620,ptmxmode=666
	232 228 0:101 / /sys ro,nosuid,nodev,noexec,relatime - sysfs sysfs ro`

	path := filepath.Join(testDir, "mountinfo")
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Error creating tempdir: %s", err)
	}
	if err := ioutil.WriteFile(path, []byte(fileContents), 0644); err != nil {
		t.Fatalf("Error writing file contents to %s: %s", path, err)
	}

	actualWhitelist, err := fileSystemWhitelist(path)
	expectedWhitelist := []string{"/kbuild", "/proc", "/dev", "/dev/pts", "/sys"}
	sort.Strings(actualWhitelist)
	sort.Strings(expectedWhitelist)
	testutil.CheckErrorAndDeepEqual(t, false, err, expectedWhitelist, actualWhitelist)
}

var tests = []struct {
	files       map[string]string
	directory   string
	expectedMap map[string][]byte
}{
	{
		files: map[string]string{
			"/workspace/foo/a": "baz1",
			"/workspace/foo/b": "baz2",
			"/kbuild/file":     "file",
		},
		directory: "/workspace/foo/",
		expectedMap: map[string][]byte{
			"workspace/foo/a": []byte("baz1"),
			"workspace/foo/b": []byte("baz2"),
			"workspace/foo":   nil,
		},
	},
	{
		files: map[string]string{
			"/workspace/foo/a": "baz1",
		},
		directory: "/workspace/foo/a",
		expectedMap: map[string][]byte{
			"workspace/foo/a": []byte("baz1"),
		},
	},
	{
		files: map[string]string{
			"/workspace/foo/a": "baz1",
			"/workspace/foo/b": "baz2",
			"/workspace/baz":   "hey",
			"/kbuild/file":     "file",
		},
		directory: "/workspace",
		expectedMap: map[string][]byte{
			"workspace/foo/a": []byte("baz1"),
			"workspace/foo/b": []byte("baz2"),
			"workspace/baz":   []byte("hey"),
			"workspace":       nil,
			"workspace/foo":   nil,
		},
	},
	{
		files: map[string]string{
			"/workspace/foo/a": "baz1",
			"/workspace/foo/b": "baz2",
			"/kbuild/file":     "file",
		},
		directory: "",
		expectedMap: map[string][]byte{
			"workspace/foo/a": []byte("baz1"),
			"workspace/foo/b": []byte("baz2"),
			"kbuild/file":     []byte("file"),
			"workspace":       nil,
			"workspace/foo":   nil,
			"kbuild":          nil,
			".":               nil,
		},
	},
}

func Test_FilesAndContents(t *testing.T) {
	for _, test := range tests {
		testDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("err setting up temp dir: %v", err)
		}
		defer os.RemoveAll(testDir)
		if err := testutil.SetupFiles(testDir, test.files); err != nil {
			t.Fatalf("err setting up files: %v", err)
		}
		fileMap, err := FilesAndContents(test.directory, testDir)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedMap, fileMap)
	}
}

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
		output, err := GetMatchedFiles(tc.srcs, tc.files)
		for _, value := range output {
			sort.Strings(value)
		}
		testutil.CheckErrorAndDeepEqual(t, false, err, tc.output, output)
	}
}
