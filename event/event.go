package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/epoch"
	"github.com/satori/go.uuid"
)

func Save(file string) error {
	rootPath, relFilePath, gtmPath, err := findPaths(file)
	if err != nil {
		return err
	}

	if err := writeEventFile(rootPath, relFilePath, gtmPath); err != nil {
		return err
	}

	return nil
}

func findPaths(file string) (string, string, string, error) {
	if !env.FileExists(file) {
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
			fmt.Sprintf("%d-%s.event", epoch.MinuteNow(), uuid.NewV4().String()[:8])),
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

func Sweep(gtmPath string, dryRun bool) (map[int64]map[string]int, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	events := make(map[int64]map[string]int, 0)
	filesToRemove := []string{}
	for _, file := range files {

		if !strings.HasSuffix(file.Name(), ".event") {
			continue
		}

		eventFilePath := filepath.Join(gtmPath, file.Name())
		filesToRemove = append(filesToRemove, eventFilePath)

		s := strings.SplitN(file.Name(), "-", 2)
		if len(s) < 2 {
			continue
		}

		fileEpoch, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			continue
		}

		_, recordedFilePath, err := readEventFile(eventFilePath)
		if err != nil {
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][recordedFilePath] += 1
	}

	if !dryRun {
		removeFiles(filesToRemove)
	}

	return events, nil
}

func removeFiles(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}
