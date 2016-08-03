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
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func newEditCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:  "edit",
		Usage: "perform an inline edit of a file from the s3 bucket",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
			cli.StringFlag{
				Name:   "e, editor",
				Usage:  "the editor to open the file with for editing",
				Value:  "vim",
				EnvVar: "EDITOR",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s"}, cmd, editFile)
		},
	}
}

//
// editFile permits an inline edit of the file
//
func editFile(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	bucket := cx.String("bucket")
	editor := cx.String("editor")

	for _, key := range cx.Args() {
		// step: retrieve the head metadata
		metadata, err := cmd.getFileMetadata(key, bucket)
		if err != nil {
			return err
		}

		// step: attempt to retrieve the data
		content, err := cmd.getFile(bucket, key)
		if err != nil {
			return fmt.Errorf("unable to retrieve keythe file: %s, error: %s", key, err)
		}

		// step: write the file to the
		path, err := inlineEdit(content, editor)
		if err != nil {
			return fmt.Errorf("unable to edit the file: %s, error: %s", key, err)
		}

		// step: upload the content to bucket
		if err := cmd.putFile(bucket, key, path, *metadata.SSEKMSKeyId); err != nil {
			os.Remove(path)
			return err
		}

		// step: add the log
		o.fields(map[string]interface{}{
			"action": "put",
			"key":    key,
			"bucket": bucket,
		}).log("successfully edited and uploaded file: s3://%s/%s\n", bucket, key)

		os.Remove(path)
	}

	return nil
}

//
// inlineEdit performs an inline edit of the file
//
func inlineEdit(content []byte, editor string) (string, error) {
	// step: create a temporary file and write the data
	tmp, err := ioutil.TempFile("/tmp", "edit.XXXXXXXX")
	if err != nil {
		return "", err
	}
	// step: write out the content of the file
	if _, err := tmp.Write(content); err != nil {
		return "", err
	}
	tmp.Close()

	// step: open the secret with the editor
	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// step: execute the editor
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return tmp.Name(), nil
}
