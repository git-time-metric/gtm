// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kilpkonn/gtm-enhanced/epoch"
	"github.com/kilpkonn/gtm-enhanced/project"
)

func pathFromSource(f string) (string, string, error) {
	if fileInfo, err := os.Stat(f); os.IsNotExist(err) || fileInfo.IsDir() {
		return "", "", project.ErrFileNotFound
	}

	repoPath, gtmPath, err := project.Paths(filepath.Dir(f))
	if err != nil {
		return "", "", err
	}

	sourcePath, err := filepath.Rel(repoPath, f)
	if err != nil {
		return "", "", err
	}

	return sourcePath, gtmPath, nil
}

func writeEventFile(sourcePath, gtmPath string) error {
	return ioutil.WriteFile(
		filepath.Join(
			gtmPath,
			fmt.Sprintf("%d.event", epoch.Now())),
		[]byte(sourcePath),
		0644)
}

func readEventFile(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.Replace(string(b), "\n", "", -1), nil
}

func removeFiles(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}
