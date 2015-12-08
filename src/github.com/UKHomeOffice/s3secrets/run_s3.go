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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func newSecretsCommand() cli.Command {
	longDescription := `Retrieve and upload secrets to Amazon S3 buckets

   Examples:
   $ s3secrets s3 put -k AWS_KMD_ID -b AWS_BUCKET -p compute/ [path/to/file|path/to/directory]
   $ s3secrets s3 get -k AWS_KMD_ID -b AWS_BUCKET [path/to/file|path/to/directory]

   # retrieve a file listing
   $ s3secrets s3 ls -l -R /
   $ s3secrets s3 ls -l -R /compute /etcd

   # read and display the file
   $ s3secrets s3 cat /compute/test.file.encrypted /somefile
   # remove files form s3 secrets bucket
   $ s3secrets s3 rm [--recursive] [path/to/file|path/to/directory]
   # list the buckets
   $ s3secrets s3 buckets [-f regex]
`
	cmd := cli.Command{
		Name:        "s3",
		Aliases:     []string{"s"},
		Usage:       "push, pull, list and cat secrets from amazon s3 secrets bucket",
		Description: longDescription,
		Subcommands: []cli.Command{
			{
				Name:  "get",
				Usage: "retrieve the secrets from the aws s3 bucket",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bucket, b",
						Usage:  "the name of the S3 bucket which the secrets",
						EnvVar: "AWS_BUCKET",
					},
					cli.StringFlag{
						Name:  "output-dir, d",
						Usage: "the path to write the descrypted secrets",
						Value: "./secrets",
					},
					cli.BoolFlag{
						Name:  "recursive, R",
						Usage: "if set, we will retrieve the files from subdirectories as well",
					},
					cli.BoolFlag{
						Name:  "no-decrypt, N",
						Usage: "if set, the file retrieved are not decrypted",
					},
				},
				Action: func(ctx *cli.Context) {
					createCommandFactory(ctx, runGetSecretsCommand)
				},
			},
			{
				Name:    "put",
				Aliases: []string{"p"},
				Usage:   "encrypt and upload secrets to the aws bucket",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "path, p",
						Usage: "the path inside the bucket where you want the encrypted files places",
					},
					cli.StringFlag{
						Name:   "kms, k",
						Usage:  "the aws kms id you wish to use to encrypt the file/s",
						EnvVar: "AWS_KMS_ID",
					},
					cli.StringFlag{
						Name:   "bucket, b",
						Usage:  "the name of the S3 bucket which the secrets",
						EnvVar: "AWS_BUCKET",
					},
				},
				Action: func(ctx *cli.Context) {
					createCommandFactory(ctx, runPutSecretsCommand)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "retrieve a listing the of files presently in the secrets bucket",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bucket, b",
						Usage:  "the name of the S3 bucket which the secrets",
						EnvVar: "AWS_BUCKET",
					},
					cli.BoolFlag{
						Name:  "recursive, R",
						Usage: "if set, the file listing will be recursive",
					},
					cli.BoolFlag{
						Name:  "long, l",
						Usage: "if set, the file listing will be detailed",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runListSecretsCommand)
				},
			},
			{
				Name:  "rm",
				Usage: "delete the file from the secrets bucket",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bucket, b",
						Usage:  "the name of the S3 bucket which the secrets",
						EnvVar: "AWS_BUCKET",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runRemoveSecretsCommand)
				},
			},
			{
				Name:  "cat",
				Usage: "retrieve and display the file to screen, decrypting if required",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "bucket, b",
						Usage:  "the name of the S3 bucket which the secrets",
						EnvVar: "AWS_BUCKET",
					},
					cli.BoolFlag{
						Name:  "no-decrypt, D",
						Usage: "if set, we will not decrypt the file regardless",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runCatSecretsCommand)
				},
			},
			{
				Name:    "buckets",
				Usage:   "list the s3 bucket in the region",
				Aliases: []string{"lsb"},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "filter, f",
						Usage: "a regex filter applied to the buckets",
						Value: ".*",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runListBucketsCommand)
				},
			},
		},
	}

	return cmd
}

// runPutSecretsCommand puts one of more secrets into the s3 bucket
func runPutSecretsCommand(cx *cli.Context, factory *commandFactory) error {
	suffix := cx.GlobalString("suffix")
	bucket := cx.String("bucket")
	bucketPath := strings.TrimSuffix(cx.String("path"), "/")
	kmsID := cx.String("kms")
	files := cx.Args()

	// step: validate the inputs
	if len(files) <= 0 {
		return fmt.Errorf("you have not specified any secrets to upload")
	}
	if bucket == "" {
		return fmt.Errorf("you have not specified a bucket to upload to")
	}

	// step: iterate the secret, encrypt and upload
	for _, path := range files {
		// check the file or directory exists
		if found := fileExists(path); !found {
			return fmt.Errorf("the file: %s does not exists", path)
		}

		uploads := []string{path}

		directory, err := isDirectory(path)
		if err != nil {
			return fmt.Errorf("failed to stat the path: %s, error: %s", path, err)
		}

		// step: if a directory, lets get a list of all the files
		if directory {
			list, err := directoryList(path)
			if err != nil {
				return fmt.Errorf("failed to get a list of files in directory: %s, error: %s", path, err)
			}
			uploads = append(uploads, list...)
		}

		// step: upload the file/s to s3
		for _, upFile := range uploads {
			fileKey := fmt.Sprintf("%s/%s%s", bucketPath, filepath.Base(upFile), suffix)

			kid := kmsID
			if strings.HasSuffix(upFile, suffix) {
				kid = ""
				log.Warnf("the file: %s is already encrypted, uploading file without encryption", upFile)
				fileKey = strings.TrimSuffix(fileKey, suffix)
			} else {
				if kmsID == "" {
					return fmt.Errorf("you have not set the kms id in order to encrypt the file")
				}
			}

			log.Debugf("pushing the file: %s, to path: %s", upFile, fileKey)

			// @TODO we could be put this back into a goroutine, but would need to wrap into multiple routines
			// to select on the wg.Wait - which i can't be arsed to do at the moment :-)
			if err := uploadFile(factory, upFile, bucket, fileKey, kid); err != nil {
				return fmt.Errorf("failed to upload the file: %s, error: %s", fileKey, err)
			}
		}
	}

	return nil
}

// runListBucketsCommand produces a list of buckets
func runListBucketsCommand(cx *cli.Context, factory *commandFactory) error {
	filter := cx.String("filter")
	filter = strings.Replace(filter, "*", ".*", -1)

	buckets, err := factory.s3.listBuckets()
	if err != nil {
		return fmt.Errorf("failed to get a list of buckets, error: %s", err)
	}

	regex, err := regexp.Compile("^" + filter + "$")
	if err != nil {
		return fmt.Errorf("invalid filter regex, error: %s", err)
	}

	for _, bucket := range buckets {
		if matched := regex.MatchString(*bucket.Name); matched {
			fmt.Printf("%-50s%s\n", *bucket.Name, bucket.CreationDate.Format("Oct 02 15:04"))
		}
	}

	return nil
}

// runGetSecretsCommand retrieves secrets from the s3 bucket and decrypts them
func runGetSecretsCommand(cx *cli.Context, factory *commandFactory) error {
	suffix := cx.GlobalString("suffix")
	bucket := cx.String("bucket")
	outputDir := cx.String("output-dir")
	recursive := cx.Bool("recursive")
	dryRun := cx.Bool("dryrun")
	noDelete := cx.Bool("no-decrypt")
	paths := cx.Args()

	if bucket == "" {
		return fmt.Errorf("you have not specified a s3 bucket name")
	}
	if len(paths) <= 0 {
		return fmt.Errorf("you have not specified any paths to retrieve")
	}

	// check the output directory exists, if not, create it
	if !fileExists(outputDir) {
		log.Warningf("the output directory: %s does not exist, attempting to create it", outputDir)
		if err := os.Mkdir(outputDir, 0700); err != nil {
			return fmt.Errorf("the specified output directory does not exist and we failed to create it, error: %s", err)
		}
	}

	for _, path := range paths {
		path = filterPath(path)

		files, err := factory.s3.listObjects(bucket, path, recursive)
		if err != nil {
			return fmt.Errorf("failed to retrieve a listing of the path: %s, error: %s", path, err)
		}

		if len(files) <= 0 {
			return fmt.Errorf("found zero files under the path: %s, were you expecting files?", path)
		}

		var filePath string

		for _, file := range files {
			// step: filter out directory entries
			if strings.HasSuffix(file.path, "/") {
				continue
			}
			// retrieve the object data
			content, err := factory.s3.getBlob(bucket, file.path)
			if err != nil {
				return fmt.Errorf("failed to retrieve the object: %s, error: %s", file.path, err)
			}
			// construct the filename
			filePath = fmt.Sprintf("%s/%s", outputDir, filepath.Base(file.path))

			if !noDelete {
				content, err = factory.kms.decrypt(content)
				if err != nil {
					return fmt.Errorf("failed to decrypt the object: %s, error: %s", file.path, err)
				}
				filePath = strings.TrimSuffix(filePath, suffix)
			}

			if err = writeFile(filePath, content, dryRun); err != nil {
				return fmt.Errorf("failed to write file: %s, error: %s", filePath, err)
			}

			log.Infof("successfully decrypted and saved the file: %s", filePath)
		}
	}

	return nil
}

// runRemoveSecretsCommand remove one of more files from the bucket
func runRemoveSecretsCommand(cx *cli.Context, factory *commandFactory) error {
	bucket := cx.String("bucket")
	paths := cx.Args()

	if bucket == "" {
		return fmt.Errorf("you have not specified a bucket to upload to")
	}

	for _, path := range paths {
		path := filterPath(path)

		err := factory.s3.removeObject(bucket, path)
		if err != nil {
			return fmt.Errorf("failed to remove the object: %s, error: %s", path, err)
		}

		log.Infof("deleted the secret: %s from the bucket", path)
	}

	return nil
}

// runListSecretsCommand retrieves a listing of the secrets under one or more paths
func runListSecretsCommand(cx *cli.Context, factory *commandFactory) error {
	bucket := cx.String("bucket")
	suffix := cx.GlobalString("suffix")
	longListing := cx.Bool("long")
	recursive := cx.Bool("recursive")
	paths := cx.Args()

	if bucket == "" {
		return fmt.Errorf("you have not specified a bucket to list files")
	}

	if len(paths) <= 0 {
		paths = append(paths, "")
	}

	// step: iterate the paths specified and list the secrets
	for _, path := range paths {
		path := filterPath(path)

		files, err := factory.s3.listObjects(bucket, path, recursive)
		if err != nil {
			return fmt.Errorf("failed to retrieve a list of files from bucket: %s, path: %s, error: %s", bucket, path, err)
		}

		fmt.Printf("total: %d\n", len(files))
		for _, file := range files {
			filename := file.path
			if strings.HasSuffix(filename, "/") {
				filename = fmt.Sprintf("%s", color.BlueString(filename))
			}
			if !strings.HasSuffix(filename, suffix) {
				filename = fmt.Sprintf("%s", color.RedString(filename))
			}

			switch longListing {
			case true:
				lastModified := "dir"
				if !file.directory {
					lastModified = file.lastModified.Format("Oct 02 15:04")
				}
				fmt.Printf("%-20s %5d %-15s  %s\n", file.owner, file.size, lastModified, filename)
			default:
				fmt.Printf("%s\n", filename)
			}
		}
	}

	return nil
}

// runCatSecretsCommand retrieves a listing of the secrets under one or more paths
func runCatSecretsCommand(cx *cli.Context, factory *commandFactory) error {
	suffix := cx.GlobalString("suffix")
	bucket := cx.String("bucket")
	noDecrypt := cx.Bool("no-decrypt")
	paths := cx.Args()

	if bucket == "" {
		return fmt.Errorf("you have not specified a bucket to upload to")
	}

	if !cx.Args().Present() {
		return fmt.Errorf("you have not specified any files to cat")
	}

	for _, path := range paths {
		var data []byte
		path := filterPath(path)

		data, err := factory.s3.getBlob(bucket, path)
		if err != nil {
			return fmt.Errorf("failed to retrieve the object: %s, error: %s", path, err)
		}

		// step: decrypt if required
		if !noDecrypt && strings.HasSuffix(path, suffix) {
			data, err = factory.kms.decrypt(data)
			if err != nil {
				return fmt.Errorf("failed to decrypt the object: %s, error: %s", path, err)
			}
		}

		os.Stdout.Write(data)
	}

	return nil
}

// uploadFile uploads the file to the s3 bucket, encrypting if required
func uploadFile(factory *commandFactory, filePath, bucket, fileKey, kmsID string) error {
	log.Debugf("uploading the file: %s to bucket: %s, path: %s", filePath, bucket, fileKey)

	// if we don't need to encrypt the file, we can only and push
	if kmsID == "" {
		fd, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fd.Close()

		if err := factory.s3.setBlob(bucket, fileKey, fd); err != nil {
			return err
		}
	} else {
		data, err := factory.kms.encryptFile(kmsID, filePath)
		if err != nil {
			return err
		}

		err = factory.s3.setBlob(bucket, fileKey, bytes.NewReader(data))
		if err != nil {
			return err
		}
	}

	log.Infof("successfully uploaded the file: %s to path: %s", filePath, fileKey)

	return nil
}

func filterPath(path string) string {
	// trim out any prefix slashes
	path = strings.TrimPrefix(path, "/")

	// switch root to nothing
	if path == "/" || path == "." {
		path = ""
	}

	return path
}
