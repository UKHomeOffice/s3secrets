/*
Copyright 2015 All rights reserved.
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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//
// hasBucket checks if the bucket exists
//
func (r cliCommand) hasBucket(bucket string) (bool, error) {
	list, err := r.listS3Buckets()
	if err != nil {
		return false, err
	}
	for _, x := range list {
		if bucket == *x.Name {
			return true, nil
		}
	}

	return false, nil
}

//
// listS3Buckets gets a list of buckets
//
func (r cliCommand) listS3Buckets() ([]*s3.Bucket, error) {
	list, err := r.s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return list.Buckets, nil
}

//
// getFileMetadata returns the head data for the specific key
//
func (r cliCommand) getFileMetadata(key, bucket string) (*s3.HeadObjectOutput, error) {
	return r.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

//
// getFile retrieves the content from a file in the bucket
//
func (r *cliCommand) getFile(bucket, key string) ([]byte, error) {
	// step: retrieve the object from the bucket
	resp, err := r.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	// step: read the content
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

//
// removeFile removes a file from a bucket
//
func (r *cliCommand) removeFile(bucket, key string) error {
	_, err := r.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}

//
// putFile uploads a file to the bucket
//
func (r *cliCommand) putFile(bucket, key, path, kmsID string) error {
	// step: open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// step: upload the file
	_, err = r.uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		Body:                 file,
		ServerSideEncryption: aws.String("aws:kms"),
		SSEKMSKeyId:          aws.String(kmsID),
	})

	return err
}

//
// listBucketKeys get all the keys from the bucket
//
func (r *cliCommand) listBucketKeys(bucket, prefix string) ([]*s3.Object, error) {
	var list []*s3.Object

	resp, err := r.s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	// step: filter out any keys which are directories
	for _, x := range resp.Contents {
		if strings.HasSuffix(*x.Key, "/") {
			continue
		}
		list = append(list, x)
	}

	return list, nil
}

//
// hasKey checks if the key exist in the bucket
//
func (r cliCommand) hasKey(key, bucket string) (bool, error) {
	keys, err := r.listBucketKeys(bucket, filepath.Dir(key))
	if err != nil {
		return false, err
	}

	for _, k := range keys {
		if key == *k.Key {
			return true, nil
		}
	}

	return false, nil
}

//
// sizeOfBucket gets the number of objects in the bucket
//
func (r cliCommand) sizeOfBucket(name string) (int, error) {
	files, err := r.listBucketKeys(name, "")
	if err != nil {
		return 0, err
	}

	return len(files), nil
}
