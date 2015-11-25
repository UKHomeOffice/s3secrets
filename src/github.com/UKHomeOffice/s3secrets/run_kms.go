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
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func newKMSCommand() cli.Command {

	longDescription := `Encrypt and decrypt a file or a collection of file's with KMS

   Note all the files will have the --suffix appended to them when encrypted

   Examples: encrypt the test.yaml file

   $ s3secrets kms encrypt -k AWS_KMD_ID test.yaml
   $ s3secrets kms decrypt -k AWS_KMD_ID path/to/directory
   $ s3secrets kms ls`

	cmd := cli.Command{
		Name:        "kms",
		Aliases:     []string{"k"},
		Usage:       "encrypting and decrypting the files with kms",
		Description: longDescription,
		Subcommands: []cli.Command{
			{
				Name:    "encrypt",
				Aliases: []string{"en"},
				Usage:   "encrypt one or more files using the KMS service",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "kms, k",
						Usage:  "the AWS KMS ID you wish to use to encrypt the file/s",
						EnvVar: "AWS_KMS_ID",
					},
					cli.BoolFlag{
						Name:  "no-delete, -N",
						Usage: "if set, post a successful encryption the original file will not be delete",
					},
					cli.BoolFlag{
						Name:  "dryrun, d",
						Usage: "if set, we performed a dryrun i.e. display the output to screen",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runKMSCommand)
				},
			},
			{
				Name:    "decrypt",
				Aliases: []string{"de"},
				Usage:   "decrypt one or more files using the KMS service",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "dryrun, d",
						Usage: "if set, we performed a dryrun i.e. display the output to screen",
					},
				},
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runKMSCommand)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list the KMS keys available to us",
				Action: func(cx *cli.Context) {
					createCommandFactory(cx, runListKMSKeysAliases)
				},
			},
		},
	}

	return cmd
}

func runListKMSKeysAliases(cx *cli.Context, factory *commandFactory) error {
	keys, err := factory.kms.list()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key.TargetKeyID != nil {
			fmt.Printf("%-40s%s\n", *key.AliasName, *key.TargetKeyID)
		}
	}

	return nil
}

func runKMSCommand(cx *cli.Context, factory *commandFactory) (err error) {
	suffix := cx.GlobalString("suffix")
	dryRun := cx.Bool("dryrun")
	kmsID := cx.String("kms")
	noDelete := cx.Bool("no-delete")
	files := cx.Args()

	encrypting := true
	if cx.Command.Name == "decrypt" {
		encrypting = false
	}

	if encrypting {
		if kmsID == "" {
			return fmt.Errorf("you must specify a kmsID in order to encrypt files")
		}
	}

	if len(files) <= 0 {
		return fmt.Errorf("you have not specified any files to act on")
	}

	for _, path := range files {
		fileList := []string{path}
		dir, err := isDirectory(path)
		if err != nil {
			return err
		}

		if dir {
			fileList, err = directoryList(path)
			if err != nil {
				return err
			}
		}

		for _, fileP := range fileList {
			var filename string
			var data []byte

			if encrypting {
				// the file should not have suffix
				if strings.HasSuffix(fileP, suffix) {
					continue
				}

				data, err = factory.kms.encryptFile(kmsID, fileP)
				if err != nil {
					return fmt.Errorf("failed to decode the file: %s, error: %s", fileP, err)
				}

				filename = fmt.Sprintf("%s%s", fileP, suffix)
			}

			if !encrypting {
				// the file should NOT have a suffix
				if !strings.HasSuffix(fileP, suffix) {
					log.Warnf("the file: %s does not have the required suffix: %s, skipping", fileP, suffix)
					continue
				}

				data, err = factory.kms.decryptFile(fileP)
				if err != nil {
					return fmt.Errorf("failed to decode the file: %s, error: %s", fileP, err)
				}

				filename = strings.TrimSuffix(fileP, suffix)
			}

			if err = writeFile(filename, data, dryRun); err != nil {
				return fmt.Errorf("failed to save the file: %s, error: %s", filename, err)
			}

			if !dryRun {
				log.Infof("successfully decrypted and saved the file: %s", filename)
			}

			if !noDelete && !dryRun {
				if err = os.Remove(fileP); err != nil {
					return fmt.Errorf("failed to delete the original file: %s, error: %s", fileP, err)
				}
			}
		}
	}

	return nil
}
