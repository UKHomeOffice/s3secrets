package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
)

// Function getS3Client returns S3 service client
func getS3Client(cfg *aws.Config) *s3.S3 {
	c := s3.New(cfg)
	return c
}

func listObjects(c *s3.S3, b string) ([]string, error) {
	l := make([]string, 0)
	resp, err := c.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b),
	})
	if err != nil {
		return l, err
	}
	for _, k := range resp.Contents {
		if strings.HasSuffix(*k.Key, fileSuffix) && *k.Size <= 5000 {
			l = append(l, *k.Key)
		}
	}
	return l, nil
}

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
