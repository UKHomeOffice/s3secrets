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
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

//
// newGetCommand creates a new get command
//
func newGetCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:  "get",
		Usage: "retrieve one or more files from the s3 bucket",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "b, bucket",
				Usage:  "the name of the s3 bucket containing the encrypted files",
				EnvVar: "AWS_S3_BUCKET",
			},
			cli.StringFlag{
				Name:  "p, perms",
				Usage: "the file permissions on any newly created files",
				Value: "0744",
			},
			cli.BoolFlag{
				Name:  "r, recursive",
				Usage: "enable recursive option and transverse all subdirectories",
			},
			cli.BoolTFlag{
				Name:  "flatten",
				Usage: "do not maintain the directory structure, flattern all files into a single directory (default true)",
			},
			cli.BoolFlag{
				Name:  "sync",
				Usage: "continously synchronize the file/s between the bucket and destination folder",
			},
			cli.DurationFlag{
				Name:  "sync-interval",
				Usage: "the time interval between successive pollings, i.e how long we should wait to recheck",
				Value: time.Duration(30 * time.Second),
			},
			cli.StringFlag{
				Name:   "d, output-dir",
				Usage:  "the path to the directory in which to save the files",
				EnvVar: "KMSCTL_OUTPUT_DIR",
				Value:  "./secrets",
			},
			cli.StringFlag{
				Name:  "f, filter",
				Usage: "apply the following regex filter to the files before retrieving",
				Value: ".*",
			},
		},
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{"l:bucket:s", "l:output-dir:s"}, cmd, getFiles)
		},
	}
}

//
// getFiles retrieve files from bucket
//
func getFiles(o *formatter, cx *cli.Context, cmd *cliCommand) (err error) {
	// step: get the inputs
	bucket := cx.String("bucket")
	directory := cx.String("output-dir")
	flatten := cx.Bool("flatten")
	recursive := cx.Bool("recursive")
	syncEnabled := cx.Bool("sync")
	syncInterval := cx.Duration("sync-interval")

	// step: validate the filter if any
	var filter *regexp.Regexp
	if filter, err = regexp.Compile(cx.String("filter")); err != nil {
		return fmt.Errorf("filter: %s is invalid, message: %s", cx.String("filter"), err)
	}

	// step: create the output directory if required
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	// step: create a signal to handle exits
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// step: create an error channel for the exiting
	errCh := make(chan error)

	go func() {
		// step: iterate the paths build a list of files were interested in
		for {
			for _, p := range getPaths(cx) {
				// step: drop the slash to for empty
				p = strings.TrimPrefix(p, "/")

				// step: list all the keys in the bucket
				files, err := cmd.listBucketKeys(bucket, p)
				if err != nil {
					errCh <- err
					return
				}

				// step: iterate the files
				for _, k := range files {
					keyName := *k.Key
					path := keyName
					filename := filepath.Base(path)

					// step: apply the filter and ignore everything were not interested in
					if !filter.MatchString(*k.Key) {
						continue
					}
					// step: are we recursive? i.e. if not, check the file ends with the filename
					if !recursive && !strings.HasSuffix(path, filename) {
						continue
					}

					// step: are we flattening the files
					switch flatten {
					case true:
						path = fmt.Sprintf("%s/%s", directory, filepath.Base(path))
					default:
						path = fmt.Sprintf("%s/%s", directory, path)
					}

					// step: process the file
					if err := processFile(path, keyName, bucket, cmd); err != nil {
						o.fields(map[string]interface{}{
							"action":      "get",
							"bucket":      bucket,
							"destination": path,
							"error":       err.Error(),
						}).log("failed to retrieve key: %s, error: %s\n", keyName, err)

						if !syncEnabled {
							errCh <- err
							return
						}
					}
					// step: add the log
					o.fields(map[string]interface{}{
						"action":      "get",
						"bucket":      bucket,
						"destination": path,
					}).log("retrieved the file: %s and wrote to: %s\n", keyName, path)
				}
			}

			// step: if not sync we can break
			if !syncEnabled {
				break
			}

			// step: otherwise we can inject delay between loops
			time.Sleep(syncInterval)
		}
		errCh <- nil
	}()

	// step: wait for an exit or signal to quit

	select {
	case <-signalCh:
	case err = <-errCh:
	}

	return err
}

//
// processFile is responsible for retrieving the files
//
func processFile(path, key, bucket string, cmd *cliCommand) error {
	// step: retrieve the file content
	content, err := cmd.getFile(bucket, key)
	if err != nil {
		return err
	}

	// step: ensure the directory structure
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// @TODO check if the files exists and only write if required

	// step: create the file for writing
	return ioutil.WriteFile(path, content, 0644)
}
