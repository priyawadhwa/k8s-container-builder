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

package integration_tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"testing"
)

var imageTests = []struct {
	description string
	repo        string
	added       []string
	deleted     []string
	modified    []string
}{
	{
		description: "test extract filesystem",
		repo:        "extract-filesystem",
		added:       []string{"/work-dir", "/work-dir/executor", "/dockerfile", "/dockerfile/Dockerfile"},
		deleted:     []string{"/proc", "/sys", "/dev", "/etc/hosts", "/etc/resolv.conf"},
	},
}

func Test_images(t *testing.T) {
	imgPrefix := "daemon://gcr.io/kbuild-test/"
	for _, test := range imageTests {
		dockerImage := imgPrefix + "docker-" + test.repo
		kbuildImage := imgPrefix + "kbuild-" + test.repo

		cmdOut, err := exec.Command("container-diff-linux-amd64", "diff", dockerImage, kbuildImage, "--type=file", "-j").Output()

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(cmdOut))

		var f interface{}
		err = json.Unmarshal(cmdOut, &f)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		adds, dels, mods := getFilenames(f)
		checkEqual(t, test.added, adds)
		checkEqual(t, test.deleted, dels)
		checkEqual(t, test.modified, mods)
	}
}

func getFilenames(f interface{}) ([]string, []string, []string) {
	diff := (f.([]interface{})[0]).(map[string]interface{})["Diff"]
	diffs := diff.(map[string]interface{})
	var adds []string
	var dels []string
	var mods []string

	addsArray := diffs["Adds"]
	if addsArray != nil {
		a := addsArray.([]interface{})
		for _, add := range a {
			filename := add.(map[string]interface{})["Name"]
			adds = append(adds, filename.(string))
		}
	}

	delsArray := diffs["Dels"]
	if delsArray != nil {
		d := delsArray.([]interface{})
		for _, del := range d {
			filename := del.(map[string]interface{})["Name"]
			dels = append(dels, filename.(string))
		}
	}

	modsArray := diffs["Mods"]
	if modsArray != nil {
		m := modsArray.([]interface{})
		for _, mod := range m {
			filename := mod.(map[string]interface{})["Name"]
			mods = append(mods, filename.(string))
		}
	}
	return adds, dels, mods
}

func checkEqual(t *testing.T, actual, expected []string) {
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%T differ.\nExpected\n%+v\nActual\n%+v", expected, expected, actual)
		return
	}
}
