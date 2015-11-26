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
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/codegangsta/cli"
)

// Version is the release version
const Version = "v0.1.4"

type commandFactory struct {
	// the s3 client interface
	s3 *s3Client
	// the kms client interface
	kms *kmsClient
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()
	app.Name = "s3secrets"
	app.Usage = "used for uploads and retrieving the secrets from the aws bucket"
	app.Version = Version
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "profile, p",
			Usage:  "the profile to use from with the aws credentials file",
			EnvVar: "AWS_DEFAULT_PROFILE",
		},
		cli.StringFlag{
			Name:  "suffix, S",
			Usage: "the suffix or extension of the file/s indicating they are encrypted",
			Value: ".encrypted",
		},
		cli.StringFlag{
			Name:   "crendentials, c",
			Usage:  "the file containing the aws credential profiles",
			EnvVar: "AWS_SHARED_CREDENTIALS_FILE",
		},
		cli.StringFlag{
			Name:   "region, R",
			Usage:  "the aws region we are speak to",
			EnvVar: "AWS_DEFAULT_REGION",
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "switch on debug / verbose logging",
		},
	}
	app.Before = func(cx *cli.Context) error {
		if cx.GlobalBool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	app.Commands = []cli.Command{
		newSecretsCommand(),
		newKMSCommand(),
	}

	app.Run(os.Args)
}

// createCommandFactory create a command factory and wraps the command processor
func createCommandFactory(cx *cli.Context, commandFunc func(*cli.Context, *commandFactory) error) {
	awsRegion := cx.GlobalString("region")
	awsProfile := cx.GlobalString("profile")
	awsCredentials := cx.GlobalString("crendentials")

	if awsRegion == "" {
		usage(cx, "you need to specify the region or export the environment variable AWS_DEFAULT_REGION")
	}

	// step: create a default aws configuration
	cfg := &aws.Config{Region: &awsRegion}

	// step: are we specifying a aws profile, if so, we need to be using the $HOME/.aws/credentials file
	if awsProfile != "" {
		cfg.Credentials = credentials.NewSharedCredentials(awsCredentials, awsProfile)
	}

	// step: create a command line factory
	factory := &commandFactory{
		s3:  newS3Client(cfg),
		kms: newKmsClient(cfg),
	}

	if err := commandFunc(cx, factory); err != nil {
		usage(cx, err.Error())
	}
}

// usage throws out an error is required and prints the usage menu
func usage(context *cli.Context, message string) {
	cli.ShowSubcommandHelp(context)
	if message != "" {
		fmt.Printf("\n[error]: %s\n", message)
		os.Exit(1)
	}

	os.Exit(0)
}
