package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/project"
)

func findPaths(file string) (string, string, string, error) {
	if fileInfo, err := os.Stat(file); os.IsNotExist(err) || fileInfo.IsDir() {
		return "", "", "", project.ErrFileNotFound
	}

	filePath, err := getFilePath(file)
	if err != nil {
		return "", "", "", err
	}

	rootPath, gtmPath, err := project.Paths(filePath)
	if err != nil {
		return "", "", "", err
	}

	relFilePath, err := filepath.Rel(rootPath, file)
	if err != nil {
		return "", "", "", err
	}

	return rootPath, relFilePath, gtmPath, nil
}

func writeEventFile(relFilePath, gtmPath string) error {
	if err := ioutil.WriteFile(
		filepath.Join(
			gtmPath,
			fmt.Sprintf("%d.event", epoch.Now())),
		[]byte(fmt.Sprintf("%s", relFilePath)),
		0644); err != nil {
		return err
	}

	return nil
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

func getFilePath(f string) (string, error) {
	p := filepath.Dir(f)
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("Unable to extract file path from %s, %s", f, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("Unable to extract file path from %s", f)
	}
	return p, nil
}
