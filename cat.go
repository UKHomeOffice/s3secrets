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
	"os"

	"github.com/urfave/cli"
)

//
// newCatCommand creates a new cat command
//
func newCatCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:  "cat",
		Usage: "retrieves and displays the contents of one or more files to the stdout",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s"}, cmd, catFiles)
		},
	}
}

//
// catFiles display one of more files to the screen
//
func catFiles(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	bucket := cx.String("bucket")

	for _, filename := range cx.Args() {
		content, err := cmd.getFile(bucket, filename)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%s", content)
	}

	return nil
}
