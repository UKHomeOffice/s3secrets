/*
Copyright 2015 All rights reserved.
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
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/urfave/cli"
)

//
// newListKMSCommand creates a new list kms key command
//
func newListKMSCommand(cmd *cliCommand) cli.Command {
	return cli.Command{
		Name:  "kms",
		Usage: "provide a listing of the kms key presently available to us",
		Action: func(cx *cli.Context) error {
			return handleCommand(cx, []string{}, cmd, listKeys)
		},
	}
}

//
// listKeys provides a listing of kms keys available
//
func listKeys(o *formatter, cx *cli.Context, cmd *cliCommand) error {
	// step: retrieve the keys from kms
	keys, err := cmd.kmsKeys()
	if err != nil {
		return err
	}

	// step: produce a listing
	for _, k := range keys {
		// step: skip any kms keys which do not have an id
		if k.TargetKeyId == nil {
			continue
		}
		o.fields(map[string]interface{}{
			"id":    *k.TargetKeyId,
			"alias": *k.AliasName,
		}).log("%-40s %-24s\n", *k.TargetKeyId, *k.AliasName)
	}

	return nil
}

//
// kmsKeys retrieves the kms keys from aws
//
func (r *cliCommand) kmsKeys() ([]*kms.AliasListEntry, error) {
	resp, err := r.kmsClient.ListAliases(&kms.ListAliasesInput{})
	if err != nil {
		return []*kms.AliasListEntry{}, err
	}

	return resp.Aliases, nil
}
