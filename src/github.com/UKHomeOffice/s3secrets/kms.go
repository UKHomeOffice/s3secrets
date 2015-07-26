package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Function getKmsClient returns KMS service client
func getKmsClient(cfg *aws.Config) *kms.KMS {
	c := kms.New(cfg)
	return c
}

func decrypt(c *kms.KMS, b *[]byte) ([]byte, error) {
	resp, err := c.Decrypt(&kms.DecryptInput{
		CiphertextBlob: *b,
	})
	if err != nil {
		return nil, err
	}
	return resp.Plaintext, nil
}
