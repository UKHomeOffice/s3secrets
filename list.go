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
	"strings"
	"time"

	"github.com/urfave/cli"
)

//
// newListCommand create's a new list command
//
func newListCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "providing a file listing of the files currently in there",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "l, long",
				Usage: "provide a detailed / long listing of the files in the bucket",
			},
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
			cli.BoolTFlag{
				Name:  "r, recursive",
				Usage: "enable recursive option and transverse all subdirectories",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s"}, cmd, listFiles)
		},
	}
}

//
// listFiles lists the files in the bucket
//
func listFiles(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	// step: get the bucket name
	bucket := cx.String("bucket")
	detailed := cx.Bool("long")
	recursive := cx.Bool("recursive")

	// step: get the paths to iterate
	for _, p := range getPaths(cx) {
		// step: get a list of paths down that path
		files, err := cmd.listBucketKeys(bucket, p)
		if err != nil {
			return err
		}

		// step: iterate the files
		for _, k := range files {
			// step: are we recursive? i.e. extract post prefix and ignore any keys which have a / in them
			if strings.Contains(strings.TrimPrefix(*k.Key, p), "/") && !recursive {
				continue
			}
			// step: are we performing a detailed listing?
			switch detailed {
			case true:
				o.fields(map[string]interface{}{
					"key":           *k.Key,
					"size":          *k.Size,
					"class":         *k.StorageClass,
					"etag":          *k.ETag,
					"owner":         *k.Owner,
					"last-modified": k.LastModified,
				}).log("%s %-10d %-20s %s\n", *k.Owner.DisplayName, *k.Size, (*k.LastModified).Format(time.RFC822), *k.Key)
			default:
				o.fields(map[string]interface{}{
					"key": *k.Key,
				}).log("%s\n", *k.Key)
			}
		}
	}

	return nil
}
