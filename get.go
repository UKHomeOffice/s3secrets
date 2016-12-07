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
func getFiles(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	var err error

	// step: get the
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
	if err = os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	// step: create a signal to handle exits and a ticker for intervals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	tickerCh := time.NewTicker(1)
	exitCh := make(chan error, 1)
	firstTime := true

	// step: create a map for etags - used to maintainer the etags of the files
	fileTags := make(map[string]string, 0)

	for {
		select {
		case err = <-exitCh:
			return err
		case <-tickerCh.C:
			if firstTime {
				tickerCh.Stop()
				tickerCh = time.NewTicker(syncInterval)
				firstTime = false
			}
			// step: iterate the paths specified on the command line
			err := func() error {
				for _, bucketPath := range getPaths(cx) {
					path := strings.TrimPrefix(bucketPath, "/")
					// step: retrieve a list of files under this path
					list, err := cmd.listBucketKeys(bucket, path)
					if err != nil {
						o.fields(map[string]interface{}{
							"bucket": bucket,
							"path":   path,
							"error":  err.Error(),
						}).log("unable to retrieve a listing in bucket: %s, path: %s\n", bucket, path)

						return err
					}

					// step: iterate the files under the path
					for _, file := range list {
						keyName := strings.TrimPrefix(*file.Key, "/")
						// step: apply the filter and ignore everything were not interested in
						if !filter.MatchString(keyName) {
							continue
						}
						// step: are we recursive? i.e. if not, check the file ends with the filename
						if !recursive && !strings.HasSuffix(path, keyName) {
							continue
						}

						// step: if we have download this file before, check the etag has changed
						if etag, found := fileTags[keyName]; found && etag == *file.ETag {
							continue // we can skip the file, nothing has changed
						}

						// step: are we flattening the files
						filename := fmt.Sprintf("%s/%s", directory, keyName)
						if flatten {
							filename = fmt.Sprintf("%s/%s", directory, filepath.Base(keyName))
						}

						// step: retrieve file and write the content to disk
						if err := processFile(filename, keyName, bucket, cmd); err != nil {
							o.fields(map[string]interface{}{
								"action":      "get",
								"bucket":      bucket,
								"destination": path,
								"error":       err.Error(),
							}).log("failed to retrieve file: %s, error: %s\n", keyName, err)

							return err
						}
						// step: update the filetags
						fileTags[keyName] = *file.ETag

						// step: add the log
						o.fields(map[string]interface{}{
							"action":      "get",
							"bucket":      bucket,
							"destination": filename,
							"etag":        file.ETag,
						}).log("retrieved the file: %s and wrote to: %s\n", keyName, filename)
					}
				}

				return nil
			}()
			// step: if we are not in a sync loop we can exit
			if !syncEnabled {
				exitCh <- err
			}
		case <-signalCh:
			o.log("exitting the synchronzition service")
			return nil
		}
	}
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
	// step: create the file for writing
	return ioutil.WriteFile(path, content, 0644)
}
