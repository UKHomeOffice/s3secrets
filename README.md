# s3secrets
Fetches, decrypts, list, removes and cats files from S3 bucket using AWS KMS

Initial idea for writing this tool was to have a way to decrypt files stored in
S3 bucket using KMS master key and write them to a tmpfs mount. s3secrets uses
KMS master key to decrypt data <=4kB in size.

Writing decrypted files to tmpfs mount is highly recommended.

## Building
You will need gb tool - http://getgb.io/.

```
git clone https://github.com/UKHomeOffice/s3secrets.git
cd s3secrets
gb build all
```

You can use docker to do the build:
```
git clone https://github.com:UKHomeOffice/s3secrets.git;
cd s3secrets
docker run --rm -it -v "$PWD":/go -w /go quay.io/ukhomeofficedigital/go-gb:1.0.0 gb build all
# or run the makefile
$ make
```
## Encrypting Data
Here is an example of how to encrypt a TLS key file using aws cli. You have to
specify which KMS master key you want to use to encrypt the data. When
decrypting the data, s3secrets does not need to know which KMS master key to
use though.

```
$ s3secrets kms encrypt -k e19da26a-dde4-4575-8f94-b840794cdb62 my_tls_key.pem
# Or just push straight s3
$ s3secrets s3 put -k e19da26a-dde4-4575-8f94-b840794cdb62 my_tls_key.pem
```

## Running
Configuration is provided via command line arguments.

```bash
$ bin/s3secrets help
NAME:
   s3secrets - used for uploads and retrieving the secrets from the aws bucket

USAGE:
   bin/s3secrets [global options] command [command options] [arguments...]

VERSION:
   0.1.2

COMMANDS:
   s3, s	push, pull, list and cat secrets from amazon s3 secrets bucket
   kms, k	encrypting and decrypting the files with kms
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --profile, -p              the profile to use from with the aws credentials file [$AWS_DEFAULT_PROFILE]
   --suffix, -S ".encrypted"  the suffix or extension of the file/s indicating they are encrypted
   --credentials, -c         the file containing the aws credential profiles [$AWS_SHARED_CREDENTIALS_FILE]
   --region, -R               the aws region we are speak to [$AWS_DEFAULT_REGION]
   --verbose, -v              if set, the sets debug logging on
   --help, -h                 show help
```

s3secrets must have access to write to the output directory. systemd.mount unit could be used to set up tmpfs mount points for different services.
The service can retrieve from multiple paths within a s3 paths and specific files

```shell
# get all files under prod/ and compute/ directories
$ s3secrets -R us-west-1 s3 get -d /etc/etcd2/tls -b my-bucket-name prod/ compute/

# get a specific file under platform/ and all of compute/
$ s3secrets -R us-west-1 s3 get -d /etc/etcd2/tls -b my-bucket-name platform/platform_ca.pem compute/
```

## Contribution
Any contribution is welcome.
