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

	"github.com/urfave/cli"
)

//
// newDeleteCommand creates a delete command
//
func newDeleteCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"rm"},
		Usage:   "Delete a file from the bucket",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s"}, cmd, deleteFile)
		},
	}
}

//
// deleteFile removes a file from the bucket
//
func deleteFile(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	if len(cx.Args()) <= 0 {
		return fmt.Errorf("you have not specified any files to delete")
	}

	bucket := cx.String("bucket")
	// step: ensure the bucket exists
	if found, err := cmd.hasBucket(bucket); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("the bucket: %s does not exist", bucket)
	}

	for _, path := range getPaths(cx) {
		if err := cmd.removeFile(bucket, path); err != nil {
			o.fields(map[string]interface{}{
				"action": "delete",
				"bucket": bucket,
				"path":   path,
				"error":  err.Error(),
			}).log("failed to remove s3://%s/%s, error: %s", bucket, path, err)
			continue
		}
		o.fields(map[string]interface{}{
			"action": "delete",
			"bucket": bucket,
			"path":   path,
		}).log("successfully deleted the file s3://%s/%s\n", bucket, path)
	}

	return nil
}
