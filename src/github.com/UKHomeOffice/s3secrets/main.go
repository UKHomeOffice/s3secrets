package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const Version = "0.1.0"

var (
	bucket       string
	region       string
	fileSuffix   string
	outputDir    string
	printVersion bool
	keyPrefix    string
)

func init() {
	flag.StringVar(&bucket, "bucket", "", "s3 bucket name with optional path")
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
	if region == "" {
		client := ec2metadata.New(nil)
		zone, err := client.GetMetadata("/placement/availability-zone")
		if err == nil {
			region = string(region[0 : len(zone)-1])
		}
	}
	var cfg *aws.Config
	if region != "" {
		cfg = &aws.Config{Region: &region}
	}

	sp := splitPath(bucket)
	bucketName := sp[0]
	if len(sp) == 2 {
		keyPrefix = sp[1]
	} else {
		keyPrefix = ""
	}

	kmsClient := getKmsClient(cfg)
	s3Client := getS3Client(cfg)
	list, err := listObjects(s3Client, bucketName, keyPrefix)
	if err != nil {
		log.Fatalln(err)
	}
	// TODO(vaijab): extract below into a goroutine
	for _, key := range list {
		blob, err := getBlob(s3Client, bucketName, key)
		if err != nil {
			fmt.Println(err)
			continue
		}
		data, err := decrypt(kmsClient, &blob)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fileName := path.Base(key)
		file := path.Join(outputDir, strings.TrimSuffix(fileName, fileSuffix))
		if err = ioutil.WriteFile(file, data, 0600); err != nil {
			fmt.Printf("Error writing to %s\n", fileName)
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

func splitPath(b string) []string {
	p := strings.SplitN(b, "/", 2)
	return p
}
