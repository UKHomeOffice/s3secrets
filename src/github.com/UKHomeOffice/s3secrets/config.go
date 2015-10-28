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
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
)

var (
	bucket       string
	region       string
	fileSuffix   string
	outputDir    string
	printVersion bool
	dryRun       bool
	paths        prefixes
)

type prefixes []string

func (p *prefixes) Set(value string) error {
	*p = append(*p, value)

	return nil
}

func (p *prefixes) String() string {
	return fmt.Sprintf("%s", *p)
}

func init() {
	flag.StringVar(&bucket, "bucket", "", "the name of the s3 bucket containing the secrets")
	flag.StringVar(&region, "region", "", "the aws region we are in")
	flag.Var(&paths, "path", "the paths you wish to retrieve from the s3 path, note you can use multiple times and can be a full filename")
	flag.StringVar(&fileSuffix, "file-suffix", ".encrypted", "the file extension on s3 files indicating encryption")
	flag.StringVar(&outputDir, "output-dir", "/run/secrets", "the path to write the descrypted secrets")
	flag.BoolVar(&printVersion, "version", false, "display the version")
	flag.BoolVar(&dryRun, "dryrun", false, "if set, perform a dry run print file content to screen")
}

// parseConfig validates the command line options
func parseConfig() error {
	flag.Parse()

	if printVersion {
		glog.Errorf("s3secrets %s\n", Version)
		os.Exit(0)
	}
	// check we have a s3 bucket
	if bucket == "" {
		return fmt.Errorf("you have not specified a s3 bucket name")
	}
	// check the output directory exists, or create it
	if !dirExists(outputDir) {
		glog.Warningf("the output directory: %s does not exist, attempting to create it", outputDir)
		if err := os.Mkdir(outputDir, 0700); err != nil {
			return fmt.Errorf("the specified output directory does not exist and we failed to create it, error: %s", err)
		}
	}

	// check we have some paths or all the default base
	if len(paths) <= 0 {
		paths.Set("")
	}

	return nil
}

func usage(message string) {
	flag.Usage()

	if message != "" {
		fmt.Printf("\n[error]: %s", message)
		os.Exit(1)
	}

	os.Exit(0)
}
