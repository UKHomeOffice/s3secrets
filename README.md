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

aws s3 cp my_tls_key.pem.encrypted s3://my-tls-keys-us-west-1
```

## Running
Configuration is provided via command line arguments.

```bash
$ s3secrets --help
Usage of s3secrets:
  -bucket string
      s3 bucket name
  -file-suffix string
      encrypted file suffix (default ".encrypted")
  -output-dir string
      output directory (default "/run/secrets")
  -region string
      aws region
  -version
      print version and exit
```

s3secrets must have access to write to the output directory. systemd.mount unit
could be used to set up tmpfs mount points for different services.

```bash
$ s3secrets --region us-west-1 --bucket my-tls-keys-us-west-1 --output-dir /etc/etcd2/tls

# should see this output
Successfully decrypted my-tls-keys-us-west-1/my_tls_key.pem.encrypted to /etc/etcd2/tls/my_tls_key.pem
```

## Known Issues
- s3secrets expects encrypted files to be stored on bucket root path.

## Contribution
Any contribution is welcome.
