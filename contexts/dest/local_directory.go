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
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/util"
	"path/filepath"
)

type LocalDirectory struct {
	root string
}

// GetFilesFromSource gets the files at path from the GCS storage bucket
func (ld *LocalDirectory) GetFilesFromSource(path string) (map[string][]byte, error) {
	// If path is an empty string, return all files at root
	if path == "" {
		return util.FilesAndContents(ld.root, ld.root)
	}
	// Otherwise, return all files at filepath.Join(root, path)
	fullPath := filepath.Join(ld.root, path)
	return util.FilesAndContents(fullPath, ld.root)
}
