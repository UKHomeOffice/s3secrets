### **S3SECRETS**

Is a command line utility for retrieving, uploading and view files encryted via the AWS KMS service.

```shell
[jest@starfury s3secrets]$ bin/s3secrets help
NAME:
   s3secrets - is a utility for interacting to s3 and kms encrypted files

USAGE:
   s3secrets [global options] command [command options] [arguments...]
   
VERSION:
   v0.0.1
   
AUTHOR(S):
   Rohith <gambol99@gmail.com> 
   
COMMANDS:
    kms		provide a listing of the kms key presently available to us
    buckets	provides a list of the buckets available to you
    list	providing a file listing of the files currently in there
    get		retrieve one or more files from the s3 bucket
    cat		retrieves and displays the contents of one or more files to the stdout
    put		upload one of more files, encrypt and place into the bucket
    edit	perform an inline edit of a file either locally or from s3 bucket

GLOBAL OPTIONS:
   -p, --profile 					the aws profile to use for static credentials [$AWS_DEFAULT_PROFILE]
   -c, --credentials "/home/jest/.aws/credentials"	the path to the credentials file container the aws profiles [$AWS_SHARED_CREDENTIALS_FILE]
   --access-key 					the aws access key to use to access the resources [$AWS_ACCESS_KEY_ID]
   --secret-key 					the aws secret key to use when accessing the resources [$AWS_SECRET_ACCESS_KEY]
   -o, --output-dir "./secrets"				the path to the directory in which to save the files [$KMSCTL_OUTPUT_DIR]
   --session-token 					the aws session token to use when accessing the resources [$AWS_SESSION_TOKEN]
   -r, --region "eu-west-1"				the aws region where the resources are located [$AWS_DEFAULT_REGION]
   -f, --format "text"					the format of the output to generate (accepts json, yaml or default text)
   --help, -h						show help
   --version, -v					print the version
```

- **Viewing the KMS keys**

```shell
[jest@starfury s3secrets]$ bin/s3secrets -p profile_name kms
74cc9f02-7795-4fe4-888e-2aae97e3eff5     alias/aws/ebs           
62c6abc6-d1d7-4203-ac3e-5733580dd4eb     alias/dev-kms-eu-west-1
75430871-d667-4fa5-bfb1-54c832f1d973     alias/prod-kms-eu-west-1
```

- **Create a bucket and upload the files**

```shell
[jest@starfury s3secrets]$ export AWS_DEFAULT_PROFILE=profile_name
[jest@starfury s3secrets]$ bin/s3secrets buckets create -n this-is-my-test-bucket-11991
successfully created the bucket: this-is-my-test-bucket-11991

[jest@starfury s3secrets]$ ls
bin  buckets.go  cmd.go  doc.go  files.go  formater.go  Godeps  keys.go  kmscli.iml  LICENSE  main.go  Makefile  release  utils.go

[jest@starfury s3secrets]$ bin/s3secrets put -k 62c6abc6-d1d7-4203-ac3e-5733580dd4eb -b this-is-my-test-bucket-11991 *.go
successfully pushed the file: buckets.go to s3://this-is-my-test-bucket-11991/buckets.go
successfully pushed the file: cmd.go to s3://this-is-my-test-bucket-11991/cmd.go
successfully pushed the file: doc.go to s3://this-is-my-test-bucket-11991/doc.go
successfully pushed the file: files.go to s3://this-is-my-test-bucket-11991/files.go
successfully pushed the file: formater.go to s3://this-is-my-test-bucket-11991/formater.go
successfully pushed the file: keys.go to s3://this-is-my-test-bucket-11991/keys.go
successfully pushed the file: main.go to s3://this-is-my-test-bucket-11991/main.go
successfully pushed the file: utils.go to s3://this-is-my-test-bucket-11991/utils.go

[jest@starfury s3secrets]$ bin/s3secrets ls -b this-is-my-test-bucket-11991 -l 
some.user 2793       26 Apr 16 13:50 UTC  buckets.go
some.user 10237      26 Apr 16 13:50 UTC  cmd.go
some.user 687        26 Apr 16 13:50 UTC  doc.go
some.user 9610       26 Apr 16 13:50 UTC  files.go
some.user 1614       26 Apr 16 13:50 UTC  formater.go
some.user 1452       26 Apr 16 13:50 UTC  keys.go
some.user 661        26 Apr 16 13:50 UTC  main.go
some.user 1445       26 Apr 16 13:50 UTC  utils.go

[jest@starfury s3secrets]$ bin/s3secrets cat -b this-is-my-test-bucket-11991 buckets.go | head
/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,

[jest@starfury s3secrets]$ bin/s3secrets buckets delete -n this-is-my-test-bucket-11991 
[error] operation failed, error: the bucket is not empty, either force (--force) deletion or empty the bucket

[jest@starfury s3secrets]$ bin/s3secrets buckets delete -n this-is-my-test-bucket-11991 --force
successfully deleted the bucket: this-is-my-test-bucket-11991
```
* **Retrieve the files from the bucket**

```shell
[jest@starfury s3secrets]$ bin/s3secrets get -b this-is-my-test-bucket-11991 -r -d ./secrets /
retrieved the file: buckets.go and wrote to: ./secrets/buckets.go
retrieved the file: cmd.go and wrote to: ./secrets/cmd.go
retrieved the file: doc.go and wrote to: ./secrets/doc.go
retrieved the file: files.go and wrote to: ./secrets/files.go
retrieved the file: formater.go and wrote to: ./secrets/formater.go
retrieved the file: keys.go and wrote to: ./secrets/keys.go
retrieved the file: main.go and wrote to: ./secrets/main.go
```
