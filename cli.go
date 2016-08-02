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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli"
)

type cliCommand struct {
	// the kms client for aws
	kmsClient *kms.KMS
	// the s3 client
	s3Client *s3.S3
	// the s3 uploader
	uploader *s3manager.Uploader
}

func newCliApplication() *cli.App {
	cmd := new(cliCommand)
	app := cli.NewApp()
	app.Name = progName
	app.Usage = "is a utility for interacting to s3 and kms encrypted files"
	app.Author = author
	app.Version = version
	app.Email = email
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "p, profile",
			Usage:  "the aws profile to use for static credentials",
			EnvVar: "AWS_DEFAULT_PROFILE",
		},
		cli.StringFlag{
			Name:   "c, credentials",
			Usage:  "the path to the credentials file container the aws profiles",
			EnvVar: "AWS_SHARED_CREDENTIALS_FILE",
			Value:  os.Getenv("HOME") + "/.aws/credentials",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "the aws access key to use to access the resources",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "the aws secret key to use when accessing the resources",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "session-token",
			Usage:  "the aws session token to use when accessing the resources",
			EnvVar: "AWS_SESSION_TOKEN",
		},
		cli.StringFlag{
			Name:  "environment-file",
			Usage: "a file containing a list of environment variables",
		},
		cli.StringFlag{
			Name:   "r, region",
			Usage:  "the aws region where the resources are located",
			EnvVar: "AWS_DEFAULT_REGION",
			Value:  "eu-west-1",
		},
		cli.StringFlag{
			Name:  "f, format",
			Usage: "the format of the output to generate (accepts json, yaml or default text)",
			Value: "text",
		},
	}

	// step: add the method for retrieving the credentials and bootstrapping
	app.Before = cmd.getCredentials()

	app.Commands = []cli.Command{
		newListKMSCommand(cmd),
		newBucketsCommand(cmd),
		newListCommand(cmd),
		newDeleteCommand(cmd),
		newCatCommand(cmd),
		newGetCommand(cmd),
		newPutCommand(cmd),
		newEditCommand(cmd),
	}

	return app
}

//
// handleCommand is a generic wrapper for handling commands, or more precisely their errors
//
func handleCommand(cx *cli.Context, options []string, cmd *cliCommand, method func(*formatter, *cli.Context, *cliCommand) error) error {
	// step: handle any panics in the command
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[error] internal error occurred, message: %s", r)
			os.Exit(1)
		}
	}()

	// step: check the required options were specified
	for _, k := range options {
		items := strings.Split(k, ":")
		if len(items) != 3 {
			panic("invalid required option definition, SCOPE:NAME:TYPE")
		}
		name := items[1]

		//
		// @Fix the cli lib IsSet does not check if the option was set by a environment variable, the
		// issue https://github.com/urfave/cli/issues/294 highlights problem. As a consequence, we can't determine
		// if the variable is actually set. The hack below attempts to remedy it.
		//
		var invalid bool

		switch scope := items[0]; scope {
		case "g":
			switch t := items[2]; t {
			case "s":
				invalid = !cx.GlobalIsSet(name) && cx.String(name) == ""
			case "a":
				invalid = !cx.GlobalIsSet(name) && len(cx.GlobalStringSlice(name)) == 0
			}
			if invalid {
				printError("the global option: '%s' is required", name)
			}
		default:
			switch t := items[2]; t {
			case "s":
				invalid = !cx.IsSet(name) && cx.String(name) == ""
			case "a":
				invalid = !cx.IsSet(name) && len(cx.StringSlice(name)) == 0
			}
			if invalid {
				printError("the command option: '%s' is required", name)
			}
		}
	}

	// step: create a cli output
	writer, err := newFormatter(cx.GlobalString("format"), os.Stdout)
	if err != nil {
		printError("error: %s", err)
	}

	// step: call the command and handle any errors
	if err := method(writer, cx, cmd); err != nil {
		printError("operation failed, error: %s", err)
	}

	return nil
}

//
// getCredentials retrieves the AWS credentials and bootstraps the cliCommand
//
func (r *cliCommand) getCredentials() func(cx *cli.Context) error {
	return func(cx *cli.Context) error {
		// step: ensure we have a region
		if cx.GlobalString("region") == "" {
			fmt.Fprintf(os.Stderr, "[error] you have not specified the aws region the resources reside\n")
			os.Exit(1)
		}
		config := &aws.Config{
			Region: aws.String(cx.GlobalString("region")),
		}

		// step: are we using static credentials
		if cx.GlobalString("access-key") != "" || cx.GlobalString("secret-ket") != "" {
			if cx.GlobalString("secret-key") == "" {
				return fmt.Errorf("you have specified a access key with a secret key")
			}
			if cx.GlobalString("access-key") == "" {
				return fmt.Errorf("you have specified a secret key with a access key")
			}
			config.Credentials = credentials.NewStaticCredentials(cx.GlobalString("access-key"),
				cx.GlobalString("secret-key"),
				cx.GlobalString("session-token"))
		} else if cx.GlobalString("profile") != "" {
			config.Credentials = credentials.NewSharedCredentials(
				cx.GlobalString("credentials"),
				cx.GlobalString("profile"))

		}

		// step: create the clients
		r.s3Client = s3.New(session.New(config))
		r.kmsClient = kms.New(session.New(config))
		r.uploader = s3manager.NewUploader(session.New(config))

		return nil
	}
}

func printError(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[error] "+message+"\n", args...)
	os.Exit(1)
}
