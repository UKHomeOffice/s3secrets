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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
)

//
// newBucketsCommand creates a new buckets command
//
func newBucketsCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:      "buckets",
		ShortName: "s3",
		Usage:     "provides a list of the buckets available to you",
		Subcommands: []cli.Command{
			{
				Name:  "ls, list",
				Usage: "retrieve a listing of all the buckets within the specified region",
				Action: func(cx *cli.Context) error {
					return handleCommand(cx, []string{}, cmd, listBuckets)
				},
			},
			{
				Name:  "create",
				Usage: "create a bucket in the specified region",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "b, bucket",
						Usage: "the name of the bucket you wish to create",
					},
				},
				Action: func(cx *cli.Context) error {
					return handleCommand(cx, []string{"l:bucket:s"}, cmd, createBucket)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "delete a bucket in the specified region",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "b, bucket",
						Usage: "the name of the bucket you wish to delete",
					},
					cli.BoolFlag{
						Name:  "force",
						Usage: "delete the bucket regardless if empty or not",
					},
				},
				Action: func(cx *cli.Context) error {
					return handleCommand(cx, []string{"l:bucket:s"}, cmd, deleteBucket)
				},
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{}, cmd, listBuckets)
		},
	}
}

func listBuckets(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	// step: get a list of buckets
	buckets, err := cmd.listS3Buckets()
	if err != nil {
		return err
	}

	// step: produce the entries
	for _, x := range buckets {
		o.fields(map[string]interface{}{
			"created": (*x.CreationDate).Format(time.RFC822Z),
			"bucket":  *x.Name,
		}).log("%-42s %20s\n", *x.Name, (*x.CreationDate).Format(time.RFC822))
	}

	return nil
}

func createBucket(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	name := cx.String("bucket")

	if found, err := cmd.hasBucket(name); err != nil {
		return err
	} else if found {
		return fmt.Errorf("the bucket already exists")
	}

	if _, err := cmd.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	}); err != nil {
		return err
	}

	o.fields(map[string]interface{}{
		"operation": "created",
		"bucket":    name,
		"created":   time.Now().Format(time.RFC822Z),
	}).log("successfully created the bucket: %s\n", name)

	return nil
}

func deleteBucket(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	name := cx.String("bucket")
	force := cx.Bool("force")

	// step: check the bucket exists
	found, err := cmd.hasBucket(name)
	if err != nil {
		return err
	} else if !found {
		return fmt.Errorf("the bucket does not exist")
	}

	// step: check if the bucket is empty
	count, err := cmd.sizeOfBucket(name)
	if err != nil {
		return err
	} else if count > 0 && !force {
		return fmt.Errorf("the bucket is not empty, either force (--force) deletion or empty the bucket")
	}

	// step: delete all the keys in the bucket first
	// @TODO find of there is a force deletion api call
	if count > 0 {
		files, err := cmd.listBucketKeys(name, "")
		if err != nil {
			return err
		}
		for _, x := range files {
			if _, err := cmd.s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(name),
				Key:    x.Key,
			}); err != nil {
				return fmt.Errorf("failed to remove the file: %s from bucket, error: %s", *x.Key, err)
			}
		}
	}
	// step: delete the bucket
	if _, err := cmd.s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}); err != nil {
		return err
	}

	o.fields(map[string]interface{}{
		"operation": "delete",
		"bucket":    name,
		"created":   time.Now().Format(time.RFC822Z),
	}).log("successfully deleted the bucket: %s\n", name)

	return nil
}
