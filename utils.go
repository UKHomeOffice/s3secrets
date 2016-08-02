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
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// getPaths returns a list of paths from the arguments, else default to base
func getPaths(cx *cli.Context) []string {
	if len(cx.Args()) <= 0 {
		return []string{""}
	}

	return cx.Args()
}

// checks if the path is a directory
func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// checks if the path is a file - kinda :-)
func isFile(path string) (bool, error) {
	dir, err := isDirectory(path)

	return !dir, err
}

// either returns the file of expands the file in the directories
func expandFiles(path string) ([]string, error) {
	var list []string
	// step: check if it's a file
	if found, err := isFile(path); err != nil {
		return list, err
	} else if found {
		return []string{path}, nil
	}
	// step: walk the files in the directories
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			list = append(list, path)
		}
		return nil
	})

	return list, err
}
