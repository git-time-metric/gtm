package event

import (
	"fmt"
	"io/ioutil"
	"os"
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

	if err := writeFile(relFilePath, gtmPath); err != nil {
		return err
	}

	return nil
}

type event struct {
	File string `json:"file"`
}

func findPaths(file string) (string, string, error) {
	if !cfg.FileExist(file) {
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

	relFilePath, err := cfg.RelativePath(file, rootPath)
	if err != nil {
		return "", "", err
	}

	return relFilePath, gtmPath, nil
}

func writeFile(relFilePath, gtmPath string) error {
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

func readFile(filePath string) (string, string, error) {
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

func Sweep(epochMarker int64, gtmPath string) (map[int64]map[string]int, error) {
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

		s := strings.SplitN(file.Name(), "-", 3)
		if len(s) < 3 {
			continue
		}

		fileEpoch, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			continue
		}

		_, recordedFilePath, err := readFile(eventFilePath)
		if err != nil {
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][recordedFilePath] += 1
	}

	remove(removeFiles)
	return events, nil
}

func remove(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}
