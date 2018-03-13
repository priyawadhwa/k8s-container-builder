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
	"path/filepath"
	"sort"
	"testing"
)

var relPathTests = []struct {
	src              string
	srcDir           bool
	filename         string
	fileDir          bool
	cwd              string
	dest             string
	expectedFilepath string
}{
	{
		src:              "context/foo",
		filename:         "context/foo",
		cwd:              "/",
		dest:             "/foo",
		expectedFilepath: "/foo",
	},
	{
		src:              "context/foo",
		filename:         "context/foo",
		cwd:              "/",
		dest:             "/foodir/",
		expectedFilepath: "/foodir/foo",
	},
	{
		src:              "foo",
		filename:         "./foo",
		cwd:              "/",
		dest:             "foo",
		expectedFilepath: "/foo",
	},
	{
		src:              "dir/",
		filename:         "dir/a/b",
		cwd:              "/",
		dest:             "pkg/",
		expectedFilepath: "/pkg/a/b",
	},
	{
		src:              "dir/",
		filename:         "dir/a/b",
		cwd:              "/newdir",
		dest:             "pkg/",
		expectedFilepath: "/newdir/pkg/a/b",
	},
	{
		src:              "./context/empty",
		srcDir:           true,
		filename:         "context/empty",
		fileDir:          true,
		cwd:              "/",
		dest:             "/empty",
		expectedFilepath: "/empty",
	},
	{
		src:              "./context/empty",
		srcDir:           true,
		filename:         "context/empty",
		fileDir:          true,
		cwd:              "/dir",
		dest:             "/empty",
		expectedFilepath: "/dir/empty",
	},
	{
		src:              "./",
		srcDir:           true,
		filename:         "./",
		fileDir:          true,
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir",
	},
	{
		src:              "./",
		srcDir:           true,
		filename:         "a",
		fileDir:          false,
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/a",
	},
	{
		src:              ".",
		srcDir:           true,
		filename:         "context/bar",
		fileDir:          false,
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/context/bar",
	},
	{
		src:              ".",
		srcDir:           true,
		filename:         "context/bar",
		fileDir:          true,
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/context/bar",
	},
}

func Test_RelativeFilepath(t *testing.T) {
	for _, test := range relPathTests {
		srcFI := testutil.MockFileInfo{
			Filename: test.src,
			Dir:      test.srcDir,
		}
		fi := testutil.MockFileInfo{
			Filename: filepath.Join("/workspace", test.filename),
			Dir:      test.fileDir,
		}
		actualFilepath, err := RelativeFilepath(test.filename, test.src, test.cwd, test.dest, srcFI, fi)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedFilepath, actualFilepath)
	}
}

var matchSourcesTests = []struct {
	srcs          []string
	files         []string
	cwd           string
	expectedFiles []string
}{
	{
		srcs: []string{
			"pkg/*",
		},
		files: []string{
			"pkg/a",
			"pkg/b",
			"/pkg/d",
			"pkg/b/d/",
			"dir/",
		},
		cwd: "/",
		expectedFiles: []string{
			"pkg/a",
			"pkg/b",
			"/pkg/d",
		},
	},
}

func Test_MatchSources(t *testing.T) {
	for _, test := range matchSourcesTests {
		actualFiles, err := MatchSources(test.srcs, test.files, test.cwd)
		sort.Strings(actualFiles)
		sort.Strings(test.expectedFiles)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedFiles, actualFiles)
	}
}
