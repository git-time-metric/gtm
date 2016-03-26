package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/epoch"
)

func findPaths(file string) (string, string, string, error) {
	if fileInfo, err := os.Stat(file); os.IsNotExist(err) || fileInfo.IsDir() {
		return "", "", "", env.ErrFileNotFound
	}

	filePath, err := env.FilePath(file)
	if err != nil {
		return "", "", "", err
	}

	rootPath, gtmPath, err := env.Paths(filePath)
	if err != nil {
		return "", "", "", err
	}

	relFilePath, err := filepath.Rel(rootPath, file)
	if err != nil {
		return "", "", "", err
	}

	return rootPath, relFilePath, gtmPath, nil
}

func writeEventFile(rootPath, relFilePath, gtmPath string) error {
	if err := ioutil.WriteFile(
		filepath.Join(
			gtmPath,
			fmt.Sprintf("%d.event", epoch.Now())),
		[]byte(fmt.Sprintf("%s,%s", rootPath, relFilePath)),
		0644); err != nil {
		return err
	}

	return nil
}

func readEventFile(filePath string) (string, string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(b), ",")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("Unable to read event %s, invalid contents %s", filePath, string(b))
	}

	return parts[0], parts[1], nil
}

func removeFiles(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}
