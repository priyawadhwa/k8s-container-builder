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
	"sort"
	"testing"
)

var testBucket = "kbuild-test-bucket"

func Test_ExtractTarFromBucket(t *testing.T) {
	testDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := UnpackTarFromGCSBucket(testBucket, testDir); err != nil {
		t.Fatalf("error unpacking tar from bucket: %v", err)
	}
	unpackedFiles, err := RelativeFiles("", testDir)
	expectedFiles := []string{
		".",
		"baz",
		"dir",
		"dir/foo",
		"dir/foo2",
	}
	sort.Strings(unpackedFiles)
	sort.Strings(expectedFiles)
	testutil.CheckErrorAndDeepEqual(t, false, err, expectedFiles, unpackedFiles)
}
