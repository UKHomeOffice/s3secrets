# s3secrets
Fetches and decrypts files from S3 bucket using AWS KMS

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
```
## Encrypting Data
Here is an example of how to encrypt a TLS key file using aws cli. You have to
specify which KMS master key you want to use to encrypt the data. When
decrypting the data, s3secrets does not need to know which KMS master key to
use though.

```
aws kms encrypt --key-id e19da26a-dde4-4575-8f94-b840794cdb62 \
  --plaintext "$(cat my_tls_key.pem)" \
  --query CiphertextBlob \
  --output text | base64 -d > my_tls_key.pem.encrypted

aws s3 cp my_tls_key.pem.encrypted s3://my-bucket-name/prod
```

## Running
Configuration is provided via command line arguments.

```bash
$ bin/s3secrets --help
Usage of bin/s3secrets:
  -alsologtostderr=false: log to standard error as well as files
  -bucket="": the name of the s3 bucket containing the secrets
  -dryrun=false: if set, perform a dry run print file content to screen
  -file-suffix=".encrypted": t	he file extension on s3 files indicating encryption
  -log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
  -log_dir="": If non-empty, write log files in this directory
  -logtostderr=false: log to standard error instead of files
  -output-dir="/run/secrets": the path to write the descrypted secrets
  -path=[]: the paths you wish to retrieve from the s3 path, note you can use multiple times and can be a full filename
  -region="": the aws region we are in
  -stderrthreshold=0: logs at or above this threshold go to stderr
  -v=0: log level for V logs
  -version=false: display the version
  -vmodule=: comma-separated list of pattern=N settings for file-filtered logging

```

s3secrets must have access to write to the output directory. systemd.mount unit could be used to set up tmpfs mount points for different services.
The service can retrieve from multiple paths within a s3 paths and specific files 

```shell
# get all files under prod/ and compute/ directories
$ s3secrets --region us-west-1 --bucket my-bucket-name -path=prod/ -path=compute/ --output-dir /etc/etcd2/tls

# get a specific file under platform/ and all of compute/
$ s3secrets --region us-west-1 --bucket my-bucket-name -path=platform/platform_ca.pem -path=compute/ --output-dir /etc/etcd2/tls
```

```bash
$ s3secrets --region us-west-1 --bucket my-bucket-name -path=prod/ --output-dir /etc/etcd2/tls

# should see this output
Successfully decrypted my-bucket-name/prod/my_tls_key.pem.encrypted to /etc/etcd2/tls/my_tls_key.pem
```

## Contribution
Any contribution is welcome.
