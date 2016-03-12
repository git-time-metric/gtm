package event

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"edgeg.io/gtm/cfg"
	"edgeg.io/gtm/epoch"
	"github.com/satori/go.uuid"
)

func Save(file string) error {
	relFilePath, gtmPath, err := findPaths(file)
	if err != nil {
		return err
	}

	if err := write(relFilePath, gtmPath); err != nil {
		return err
	}

	return nil
}

func findPaths(file string) (string, string, error) {
	if !cfg.FileExists(file) {
		return "", "", cfg.ErrFileNotFound
	}

	filePath, err := cfg.FilePath(file)
	if err != nil {
		return "", "", err
	}

	rootPath, gtmPath, err := cfg.Paths(filePath)
	if err != nil {
		return "", "", err
	}

	relFilePath, err := filepath.Rel(rootPath, file)
	if err != nil {
		return "", "", err
	}

	return relFilePath, gtmPath, nil
}

func write(relFilePath, gtmPath string) error {
	if err := ioutil.WriteFile(
		filepath.Join(
			gtmPath,
			fmt.Sprintf("%d-%s.event", epoch.MinuteNow(), uuid.NewV4().String()[:8])),
		[]byte(fmt.Sprintf("%s,%s", gtmPath, relFilePath)),
		0644); err != nil {
		return err
	}

	return nil
}

func read(filePath string) (string, string, error) {
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

func Sweep(gtmPath string) (map[int64]map[string]int, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	events := make(map[int64]map[string]int, 0)
	removeFiles := []string{}
	for _, file := range files {

		if !strings.HasSuffix(file.Name(), ".event") {
			continue
		}

		eventFilePath := filepath.Join(gtmPath, file.Name())
		removeFiles = append(removeFiles, eventFilePath)

		s := strings.SplitN(file.Name(), "-", 2)
		if len(s) < 2 {
			continue
		}

		fileEpoch, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			continue
		}

		_, recordedFilePath, err := read(eventFilePath)
		if err != nil {
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][recordedFilePath] += 1
	}

	cfg.RemoveFiles(removeFiles)
	return events, nil
}
