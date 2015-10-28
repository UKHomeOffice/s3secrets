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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// getS3Client returns S3 service client
func getS3Client(cfg *aws.Config) *s3.S3 {
	return s3.New(cfg)
}

// listObjects returns a list of entries in a bucket path
func listObjects(c *s3.S3, b, p string) ([]string, error) {
	var l []string
	resp, err := c.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b),
		Prefix: aws.String(p),
	})
	if err != nil {
		return l, err
	}
	for _, k := range resp.Contents {
		// step: do NOT allow base to propagate
		if p == "" && strings.Contains(*k.Key, "/") {
			continue
		}
		// step: is MUST start with out prefix
		if !strings.HasPrefix(*k.Key, p) {
			continue
		}
		if strings.HasSuffix(*k.Key, fileSuffix) && *k.Size <= 5000 {
			l = append(l, *k.Key)
		}
	}
	return l, nil
}

// getBlob return the data associated to the entry
func getBlob(c *s3.S3, b, k string) ([]byte, error) {
	resp, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	})
	if err != nil {
		return nil, err
	}
	blob := make([]byte, *resp.ContentLength)
	_, err = resp.Body.Read(blob)
	return blob, nil
}
