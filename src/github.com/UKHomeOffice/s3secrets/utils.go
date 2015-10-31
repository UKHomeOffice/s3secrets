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
	"os"
	"path/filepath"
	"syscall"
)

// isDirectory checks if the path is a directory
func isDirectory(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

// directoryList returns a listing of the files in a directory
func directoryList(path string) ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return fileList, err
	}

	return fileList, nil
}

// fileExists checks if the directory exists
func fileExists(f string) bool {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return true
	}

	return true
}

// writeFile writes the content to the stdout or the file
func writeFile(path string, content []byte, dryRun bool) (err error) {
	var file *os.File

	if dryRun {
		file = os.Stdout
	} else {
		file, err = os.OpenFile(path, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	if _, err = file.Write(content); err != nil {
		return err
	}

	return nil
}
