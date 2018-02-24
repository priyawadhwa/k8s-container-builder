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

package dest

import (
	"archive/tar"
	"github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
)

type TarContext struct {
	filePath string
}

// GetFilesFromSource gets the files at path from the tarball
func (t TarContext) GetFilesFromSource(path string) (map[string][]byte, error) {
	logrus.Infof("Reading from %s", t.filePath)
	file, err := os.Open(t.filePath)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	reader := tar.NewReader(file)
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if util.HasFilepathPrefix(hdr.Name, path) || path == "" {
			fileContents, err := ioutil.ReadAll(reader)
			logrus.Debugf("Getting %s from tar source", hdr.Name)
			if err != nil {
				return nil, err
			}
			files[hdr.Name] = fileContents
		}
	}
	return files, nil
}
