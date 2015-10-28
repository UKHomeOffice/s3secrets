/*
Copyright 2015 HomeOffice All rights reserved.

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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"
)

// Version is the version of the application
const Version = "0.1.1"

func main() {
	var cfg *aws.Config

	if err := parseConfig(); err != nil {
		usage(err.Error())
	}

	if region != "" {
		cfg = &aws.Config{Region: region}
	}

	kmsClient := getKmsClient(cfg)
	s3Client := getS3Client(cfg)

	// step: iterate the paths in the bucket
	for _, path := range paths {
		glog.V(3).Infof("retrieving a listing of files from bucket: %s, path: %s", bucket, path)
		// get a list of files
		list, err := listObjects(s3Client, bucket, path)
		if err != nil {
			glog.Fatalf("failed to retrieve a listing of the path: %s, error: %s", path, err)
		}

		glog.V(3).Infof("found %d files bucket: %s, path: %s", len(list), bucket, path)

		// step: iterate the files, retrieve, decrypt and save
		for _, filename := range list {
			filePath := fmt.Sprintf("%s/%s/%s", bucket, path, filename)
			glog.V(3).Infof("attemping to fetch the file: %s from s3", filePath)

			blob, err := getBlob(s3Client, bucket, filename)
			if err != nil {
				glog.Errorf("failed to retrieve the file: %s, error: %s", filePath, err)
				continue
			}

			glog.V(3).Infof("decrypting the file: %s", filePath)

			data, err := decrypt(kmsClient, &blob)
			if err != nil {
				glog.Errorf("failed to decrypt the file: %s, error: %s", filePath, err)
				continue
			}

			if err = writeFile(filename, data); err != nil {
				glog.Errorf("failed to write file: %s, error: %s", filePath, err)
				continue
			}

			glog.V(3).Infof("successfully decrypted and saved the file: %s", filePath)
		}
	}
}

// writeFile writes the content to the stdout or the file
func writeFile(filename string, content []byte) (err error) {
	var file *os.File

	filePath := fmt.Sprintf("%s/%s", outputDir, strings.TrimSuffix(filepath.Base(filename), fileSuffix))

	if dryRun {
		file = os.Stdout
	} else {
		file, err = os.OpenFile(filePath, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	if _, err = file.Write(content); err != nil {
		return err
	}

	return nil
}

// dirExists checks if the directory exists
func dirExists(f string) bool {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return true
	}

	return true
}
