package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const Version = "0.0.1"

var (
	bucket       string
	region       string
	fileSuffix   string
	outputDir    string
	printVersion bool
)

func init() {
	flag.StringVar(&bucket, "bucket", "", "s3 bucket name")
	flag.StringVar(&region, "region", "", "aws region")
	flag.StringVar(&fileSuffix, "file-suffix", ".encrypted", "encrypted file suffix")
	flag.StringVar(&outputDir, "output-dir", "/run/secrets", "output directory")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
}

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("s3secrets %s\n", Version)
		os.Exit(0)
	}
	if bucket == "" {
		fmt.Println("Please specify S3 bucket name. Exiting..")
		os.Exit(1)
	}
	if !dirExists(outputDir) {
		fmt.Println("Output dir does not exist. Exiting..")
		os.Exit(1)
	}
	var cfg *aws.Config
	if region != "" {
		cfg = &aws.Config{Region: region}
	}
	s3Client := getS3Client(cfg)
	list, err := listObjects(s3Client, bucket)
	if err != nil {
		log.Fatalln(err)
	}
	kmsClient := getKmsClient(cfg)
	// TODO(vaijab): extract below into a goroutine
	for _, key := range list {
		blob, err := getBlob(s3Client, bucket, key)
		if err != nil {
			fmt.Println(err)
		}
		data, err := decrypt(kmsClient, &blob)
		if err != nil {
			fmt.Println(err)
		}
		file := path.Join(outputDir, strings.TrimSuffix(key, fileSuffix))
		if err = ioutil.WriteFile(file, data, 0600); err != nil {
			fmt.Printf("Error writing to %s\n", file)
		} else {
			fmt.Printf("Successfully decrypted %s to %s\n", path.Join(bucket, key), file)
		}
	}
}

func dirExists(f string) bool {
	if _, err := os.Stat(f); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return true
		}
	}
	return true
}
