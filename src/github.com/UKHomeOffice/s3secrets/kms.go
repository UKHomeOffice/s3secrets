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
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

type kmsClient struct {
	// the client to speak to the kms service
	client *kms.KMS
}

// newKmsClient returns KMS service client
func newKmsClient(cfg *aws.Config) *kmsClient {
	return &kmsClient{
		client: kms.New(cfg),
	}
}

// encrypt encodes the plaintext block into encrypted form
func (r *kmsClient) encrypt(kmsID string, plain []byte) ([]byte, error) {
	resp, err := r.client.Encrypt(&kms.EncryptInput{
		KeyID:     &kmsID,
		Plaintext: plain,
	})
	if err != nil {
		return []byte(""), err
	}

	return resp.CiphertextBlob, nil
}

// decryptFile reads in and decodes a file
func (r *kmsClient) decryptFile(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte(""), err
	}

	data, err := r.decrypt(content)
	if err != nil {
		return []byte(""), err
	}

	return data, nil
}

// decrypt decodes the file and returns the byte stream
func (r *kmsClient) decrypt(b []byte) ([]byte, error) {
	resp, err := r.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: b,
	})
	if err != nil {
		return nil, err
	}

	return resp.Plaintext, nil
}

// encryptFile reads in and encrypts the data from a file
func (r *kmsClient) encryptFile(kmsID, path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte(""), err
	}

	data, err := r.encrypt(kmsID, content)
	if err != nil {
		return []byte(""), err
	}

	return data, nil
}

// list retrieves a list of kms aliases
func (r *kmsClient) list() ([]*kms.AliasListEntry, error) {
	resp, err := r.client.ListAliases(&kms.ListAliasesInput{})
	if err != nil {
		return []*kms.AliasListEntry{}, err
	}

	return resp.Aliases, nil
}
