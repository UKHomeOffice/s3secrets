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
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// s3Client is the client interface to the s3 objects and buckets
type s3Client struct {
	region string
	// the client for s3
	client *s3.S3
	// the s3 uploader
	uploader *s3manager.Uploader
}

type objectFile struct {
	path string
	// the size of the object
	size int64
	// the owner of the file
	owner string
	// a directory?
	directory bool
	// last modified
	lastModified *time.Time
}

// newS3Client returns S3 service client
func newS3Client(cfg *aws.Config) *s3Client {
	client := s3.New(cfg)
	uploadOpts := s3manager.DefaultUploadOptions
	uploadOpts.S3 = client
	uploader := s3manager.NewUploader(uploadOpts)

	return &s3Client{
		region:   cfg.Region,
		client:   client,
		uploader: uploader,
	}
}

// listObjects returns a list of entries in a bucket
func (r *s3Client) listObjects(bucket, path string, recursive bool) (list []*objectFile, err error) {

	log.Debugf("retrieving the secrets from bucket: %s, path: '%s'", bucket, path)

	pathBaseDirectory := filepath.Dir(path)
	if pathBaseDirectory == "." {
		pathBaseDirectory = ""
	}

	var regex *regexp.Regexp
	switch recursive {
	case true:
		regex, err = regexp.Compile("^" + strings.Replace(path, "*", ".*", 0) + "[\\w-_\\.]*[$/]?.*")
	default:
		regex, err = regexp.Compile("^" + strings.Replace(path, "*", ".*", 0) + "[\\w-_\\.]*[$/]?")
	}
	if err != nil {
		return list, err
	}

	resp, err := r.client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(pathBaseDirectory),
	})
	if err != nil {
		return list, err
	}

	seenDir := make(map[string]bool, 0)

	for _, object := range resp.Contents {
		matches := regex.FindAllString(*object.Key, 1)
		if len(matches) <= 0 {
			continue
		}

		entry := &objectFile{path: matches[0]}
		if !strings.HasSuffix(entry.path, "/") {
			entry.size = *object.Size
			entry.lastModified = object.LastModified
			entry.owner = *object.Owner.DisplayName
		} else {
			if found := seenDir[entry.path]; found {
				continue
			}
			seenDir[entry.path] = true
			entry.directory = true
		}

		list = append(list, entry)
	}

	return list, nil
}

// listBucket returns a bucket in the region
func (r *s3Client) listBuckets() ([]*s3.Bucket, error) {
	resp, err := r.client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return []*s3.Bucket{}, err
	}

	return resp.Buckets, nil
}

// setBlob uploads a blob to s3 bucket
func (r *s3Client) setBlob(bucket, path string, reader io.Reader) error {
	log.Debugf("putting file into bucket: %s, path: %s", bucket, path)
	if _, err := r.uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: &bucket,
		Key:    &path,
	}); err != nil {
		return err
	}

	return nil
}

// getBlob return the data associated to the entry
func (r *s3Client) getBlob(bucket, path string) ([]byte, error) {
	log.Debugf("attempting to retrieve the object, bucket: %s, path: %s", bucket, path)
	resp, err := r.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("/" + path),
	})
	if err != nil {
		return nil, err
	}

	blob := make([]byte, *resp.ContentLength)
	_, err = resp.Body.Read(blob)

	return blob, nil
}

// removeObject deletes an object from s3
func (r *s3Client) removeObject(bucket, path string) error {
	log.Debugf("attmepting to remove the object: bucket: %s, path: %s", bucket, path)

	if _, err := r.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &path,
	}); err != nil {
		return err
	}

	return nil
}
