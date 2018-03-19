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
package storage

import (
	"github.com/GoogleCloudPlatform/k8s-container-builder/testutil"
	"testing"
)

// This test bucket can be found in the kbuild-test project
var bucketName = "kbuild-test-bucket"

var tests = []struct {
	path        string
	expectedMap map[string][]byte
}{
	{
		path: "",
		expectedMap: map[string][]byte{
			"foo":        []byte("foo"),
			"bat/baz":    []byte("test"),
			"empty_dir/": {},
			"empty_file": {},
		},
	},
	{
		path: "foo",
		expectedMap: map[string][]byte{
			"foo": []byte("foo"),
		},
	},
	{
		path: "bat/",
		expectedMap: map[string][]byte{
			"bat/baz": []byte("test"),
		},
	},
	{
		path: "bat",
		expectedMap: map[string][]byte{
			"bat/baz": []byte("test"),
		},
	},
	{
		path:        "empty/",
		expectedMap: map[string][]byte{},
	},
}

func TestStorage(t *testing.T) {
	for _, test := range tests {
		bucketFilesMap, err := GetFilesFromStorageBucket(bucketName, test.path)
		testutil.CheckErrorAndDeepEqual(t, false, err, test.expectedMap, bucketFilesMap)
	}
}
