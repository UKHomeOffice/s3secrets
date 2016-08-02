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
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

func newPutCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:  "put",
		Usage: "upload one of more files, encrypt and place into the bucket",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
			cli.StringFlag{
				Name:   "k, kms",
				Usage:  "the aws kms id to use when performing operations",
				EnvVar: "AWS_KMS_ID",
			},
			cli.StringFlag{
				Name:  "p, path",
				Usage: "use this are the path inside the bucket, rather than the path to the file",
			},
			cli.BoolFlag{
				Name:  "flatten",
				Usage: "do not maintain the directory structure, flatten all files into a single directory",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s", "l:kms:s"}, cmd, putFiles)
		},
	}
}

//
// putFiles uploads a selection of files into the bucket
//
func putFiles(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	bucket := cx.String("bucket")
	kms := cx.String("kms")
	flatten := cx.Bool("flatten")
	path := cx.String("path")

	if flatten && path != "" {
		return fmt.Errorf("invalid option, you cannot flatten *and* specify a path")
	}

	// step: ensure the bucket exists
	if found, err := cmd.hasBucket(bucket); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("the bucket: %s does not exist", bucket)
	}

	// check: we need any least one argument
	if len(cx.Args()) <= 0 {
		return fmt.Errorf("you have not specified any files to upload")
	}

	// step: iterate the paths and upload the files
	for _, p := range getPaths(cx) {
		// step: get a list of files under this path
		files, err := expandFiles(p)
		if err != nil {
			return fmt.Errorf("failed to process path: %s, error: %s", p, err)
		}
		// step: iterate the files in the path
		for _, filename := range files {
			// step: construct the key for this file
			keyName := filename
			if flatten {
				keyName = filepath.Base(keyName)
			}
			if path != "" {
				keyName = fmt.Sprintf("%s/%s", strings.TrimRight(path, "/"), filepath.Base(keyName))
			}

			// step: upload the file to the bucket
			if err := cmd.putFile(bucket, keyName, filename, kms); err != nil {
				return fmt.Errorf("failed to put the file: %s, error: %s", filename, err)
			}

			// step: add the log
			o.fields(map[string]interface{}{
				"action": "put",
				"path":   filename,
				"bucket": bucket,
				"key":    keyName,
			}).log("successfully pushed the file: %s to s3://%s/%s\n", filename, bucket, keyName)
		}
	}

	return nil
}
