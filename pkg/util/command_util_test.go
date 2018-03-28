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
	"sort"
	"testing"
)

var testUrl = "https://github.com/GoogleCloudPlatform/runtimes-common/blob/master/LICENSE"

var testEnvReplacement = []struct {
	path         string
	command      string
	envs         []string
	isFilepath   bool
	expectedPath string
}{
	{
		path:    "/simple/path",
		command: "WORKDIR /simple/path",
		envs: []string{
			"simple=/path/",
		},
		isFilepath:   true,
		expectedPath: "/simple/path",
	},
	{
		path:    "/simple/path/",
		command: "WORKDIR /simple/path/",
		envs: []string{
			"simple=/path/",
		},
		isFilepath:   true,
		expectedPath: "/simple/path/",
	},
	{
		path:    "${a}/b",
		command: "WORKDIR ${a}/b",
		envs: []string{
			"a=/path/",
			"b=/path2/",
		},
		isFilepath:   true,
		expectedPath: "/path/b",
	},
	{
		path:    "/$a/b",
		command: "COPY ${a}/b /c/",
		envs: []string{
			"a=/path/",
			"b=/path2/",
		},
		isFilepath:   true,
		expectedPath: "/path/b",
	},
	{
		path:    "/$a/b/",
		command: "COPY /${a}/b /c/",
		envs: []string{
			"a=/path/",
			"b=/path2/",
		},
		isFilepath:   true,
		expectedPath: "/path/b/",
	},
	{
		path:    "\\$foo",
		command: "COPY \\$foo /quux",
		envs: []string{
			"foo=/path/",
		},
		isFilepath:   true,
		expectedPath: "$foo",
	},
	{
		path:    "8080/$protocol",
		command: "EXPOSE 8080/$protocol",
		envs: []string{
			"protocol=udp",
		},
		expectedPath: "8080/udp",
	},
}

func Test_EnvReplacement(t *testing.T) {
	for _, test := range testEnvReplacement {
		actualPath, err := ResolveEnvironmentReplacement(test.path, test.envs, test.isFilepath)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedPath, actualPath)

	}
}

var buildContextPath = "../../integration_tests/"

var destinationFilepathTests = []struct {
	srcName          string
	filename         string
	dest             string
	cwd              string
	buildcontext     string
	expectedFilepath string
}{
	{
		srcName:          "context/foo",
		filename:         "context/foo",
		dest:             "/foo",
		cwd:              "/",
		expectedFilepath: "/foo",
	},
	{
		srcName:          "context/foo",
		filename:         "context/foo",
		dest:             "/foodir/",
		cwd:              "/",
		expectedFilepath: "/foodir/foo",
	},
	{
		srcName:          "context/foo",
		filename:         "./context/foo",
		cwd:              "/",
		dest:             "foo",
		expectedFilepath: "/foo",
	},
	{
		srcName:          "context/bar/",
		filename:         "context/bar/bam/bat",
		cwd:              "/",
		dest:             "pkg/",
		expectedFilepath: "/pkg/bam/bat",
	},
	{
		srcName:          "context/bar/",
		filename:         "context/bar/bam/bat",
		cwd:              "/newdir",
		dest:             "pkg/",
		expectedFilepath: "/newdir/pkg/bam/bat",
	},
	{
		srcName:          "./context/empty",
		filename:         "context/empty",
		cwd:              "/",
		dest:             "/empty",
		expectedFilepath: "/empty",
	},
	{
		srcName:          "./context/empty",
		filename:         "context/empty",
		cwd:              "/dir",
		dest:             "/empty",
		expectedFilepath: "/empty",
	},
	{
		srcName:          "./",
		filename:         "./",
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir",
	},
	{
		srcName:          "./",
		filename:         "context/foo",
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/context/foo",
	},
	{
		srcName:          ".",
		filename:         "context/bar",
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/context/bar",
	},
	{
		srcName:          ".",
		filename:         "context/bar",
		cwd:              "/",
		dest:             "/dir",
		expectedFilepath: "/dir/context/bar",
	},
	{
		srcName:          "context/foo",
		filename:         "context/foo",
		cwd:              "/test",
		dest:             ".",
		expectedFilepath: "/test/foo",
	},
}

func Test_DestinationFilepath(t *testing.T) {
	for _, test := range destinationFilepathTests {
		actualFilepath, err := DestinationFilepath(test.filename, test.srcName, test.dest, test.cwd, buildContextPath)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedFilepath, actualFilepath)
	}
}

var urlDestFilepathTests = []struct {
	url          string
	cwd          string
	dest         string
	expectedDest string
}{
	{
		url:          "https://something/something",
		cwd:          "/test",
		dest:         ".",
		expectedDest: "/test/something",
	},
	{
		url:          "https://something/something",
		cwd:          "/cwd",
		dest:         "/test",
		expectedDest: "/test",
	},
	{
		url:          "https://something/something",
		cwd:          "/test",
		dest:         "/dest/",
		expectedDest: "/dest/something",
	},
}

func Test_UrlDestFilepath(t *testing.T) {
	for _, test := range urlDestFilepathTests {
		actualDest := URLDestinationFilepath(test.url, test.dest, test.cwd)
		testutil.CheckErrorAndDeepEqual(t, false, nil, test.expectedDest, actualDest)
	}
}

var matchSourcesTests = []struct {
	srcs          []string
	files         []string
	expectedFiles []string
}{
	{
		srcs: []string{
			"pkg/*",
			testUrl,
		},
		files: []string{
			"pkg/a",
			"pkg/b",
			"/pkg/d",
			"pkg/b/d/",
			"dir/",
		},
		expectedFiles: []string{
			"pkg/a",
			"pkg/b",
			testUrl,
		},
	},
}

func Test_MatchSources(t *testing.T) {
	for _, test := range matchSourcesTests {
		actualFiles, err := matchSources(test.srcs, test.files)
		sort.Strings(actualFiles)
		sort.Strings(test.expectedFiles)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedFiles, actualFiles)
	}
}

var isSrcValidTests = []struct {
	srcsAndDest []string
	files       map[string][]string
	shouldErr   bool
}{
	{
		srcsAndDest: []string{
			"src1",
			"src2",
			"dest",
		},
		files: map[string][]string{
			"src1": {
				"file1",
			},
			"src2:": {
				"file2",
			},
		},
		shouldErr: true,
	},
	{
		srcsAndDest: []string{
			"src1",
			"src2",
			"dest/",
		},
		files: map[string][]string{
			"src1": {
				"file1",
			},
			"src2:": {
				"file2",
			},
		},
		shouldErr: false,
	},
	{
		srcsAndDest: []string{
			"src2/",
			"dest",
		},
		files: map[string][]string{
			"src1": {
				"file1",
			},
			"src2:": {
				"file2",
			},
		},
		shouldErr: false,
	},
	{
		srcsAndDest: []string{
			"src2",
			"dest",
		},
		files: map[string][]string{
			"src1": {
				"file1",
			},
			"src2:": {
				"file2",
			},
		},
		shouldErr: false,
	},
	{
		srcsAndDest: []string{
			"src2",
			"src*",
			"dest/",
		},
		files: map[string][]string{
			"src1": {
				"file1",
			},
			"src2:": {
				"file2",
			},
		},
		shouldErr: false,
	},
	{
		srcsAndDest: []string{
			"src2",
			"src*",
			"dest",
		},
		files: map[string][]string{
			"src2": {
				"src2/a",
				"src2/b",
			},
			"src*": {},
		},
		shouldErr: true,
	},
	{
		srcsAndDest: []string{
			"src2",
			"src*",
			"dest",
		},
		files: map[string][]string{
			"src2": {
				"src2/a",
			},
			"src*": {},
		},
		shouldErr: false,
	},
	{
		srcsAndDest: []string{
			"src2",
			"src*",
			"dest",
		},
		files: map[string][]string{
			"src2": {},
			"src*": {},
		},
		shouldErr: true,
	},
}

func Test_IsSrcsValid(t *testing.T) {
	for _, test := range isSrcValidTests {
		err := IsSrcsValid(test.srcsAndDest, test.files)
		testutil.CheckError(t, test.shouldErr, err)
	}
}

var testResolveSources = []struct {
	srcsAndDest []string
	expectedMap map[string][]string
}{
	{
		srcsAndDest: []string{
			"context/foo",
			"context/b*",
			testUrl,
			"dest/",
		},
		expectedMap: map[string][]string{
			"context/foo": {
				"context/foo",
			},
			"context/bar": {
				"context/bar",
				"context/bar/bam",
				"context/bar/bam/bat",
				"context/bar/bat",
				"context/bar/baz",
			},
			testUrl: {
				testUrl,
			},
		},
	},
}

func Test_ResolveSources(t *testing.T) {
	for _, test := range testResolveSources {
		actualMap, err := ResolveSources(test.srcsAndDest, buildContextPath)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedMap, actualMap)
	}
}

var testRemoteUrls = []struct {
	url   string
	valid bool
}{
	{
		url:   testUrl,
		valid: true,
	},
	{
		url:   "not/real/",
		valid: false,
	},
	{
		url:   "https://url.com/something/not/real",
		valid: false,
	},
}

func Test_RemoteUrls(t *testing.T) {
	for _, test := range testRemoteUrls {
		valid := IsSrcRemoteFileURL(test.url)
		testutil.CheckErrorAndDeepEqual(t, false, nil, test.valid, valid)
	}

}
