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
	"cloud.google.com/go/storage"
	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/k8s-container-builder/pkg/constants"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"path/filepath"
)

// UnpackTarFromGCSBucket unpacks the kbuild.tar file in the given bucket to the given directory
func UnpackTarFromGCSBucket(bucketName, directory string) error {
	// Get the tar from the bucket
	tarPath, err := getTarFromBucket(bucketName, directory)
	if err != nil {
		return err
	}
	logrus.Debugf("Unpacking source context from %s...", tarPath)
	// Now, unpack the tar to a build context, and return the path to the build context
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	if err := pkgutil.UnTar(file, directory, nil); err != nil {
		return err
	}
	// Remove the tar so it doesn't interfere with subsequent commands
	logrus.Debugf("Deleting %s", tarPath)
	return os.Remove(tarPath)
}

// getTarFromBucket gets kbuild.tar from the GCS bucket and saves it to the filesystem
// It returns the path to the tar file
func getTarFromBucket(bucketName, directory string) (string, error) {
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
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	tarPath := filepath.Join(directory, constants.KbuildTar)
	if err := CreateFile(tarPath, contents, 0600); err != nil {
		return "", err
	}
	logrus.Debugf("Copied tarball %s from GCS bucket %s to %s", constants.KbuildTar, bucketName, tarPath)
	return tarPath, nil
}
