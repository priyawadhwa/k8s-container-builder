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
	"archive/tar"
	"cloud.google.com/go/storage"
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
)

// UnpackTarFromGCSBucket unpacks the kbuild.tar file in the given bucket at the given directory
func UnpackTarFromGCSBucket(bucketName, directory string) error {
	// Get the tar from the bucket
	tarPath, err := getTarFromBucket(bucketName)
	if err != nil {
		return err
	}

	// Now, unpack the tar to a build context, and return the path to the build context
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	return pkgutil.UnTar(file, directory, nil)
}

// getTarFromBucket gets kbuild.tar from the GCS bucket and saves it to the filesystem
// It returns the path to the tar file
func getTarFromBucket(bucketName string) (string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	bucket := client.Bucket(bucketName)

	// Get the tarfile kbuild.tar from the GCS bucket, and save it to a tar object

	reader, err := bucket.Object(constants.KbuildTar).NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	tarPath := filepath.Join(constants.KbuildDir, constants.KbuildTar)
	f, err := os.Create(tarPath)
	if err != nil {
		return "", nil
	}
	logrus.Debugf("Copied tarball %s from GCS bucket %s to %s", constants.KbuildTar, bucketName, tarPath)
	defer f.Close()

	w := tar.NewWriter(f)
	defer w.Close()

	_, err = io.Copy(w, reader)
	return tarPath, err
}
