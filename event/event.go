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

func Sweep(gtmPath string, dryRun bool) (map[int64]map[string]int, error) {
	events := make(map[int64]map[string]int, 0)

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return events, err
	}

	filesToRemove := []string{}
	var prevEpoch int64
	var prevFilePath string
	for i := range files {

		if !strings.HasSuffix(files[i].Name(), ".event") {
			continue
		}

		eventFilePath := filepath.Join(gtmPath, files[i].Name())
		filesToRemove = append(filesToRemove, eventFilePath)

		s := strings.SplitN(files[i].Name(), ".", 2)
		if len(s) != 2 {
			continue
		}

		fileEpoch, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			continue
		}
		fileEpoch = epoch.Minute(fileEpoch)

		_, filePath, err := readEventFile(eventFilePath)
		if err != nil {
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][filePath]++

		// Add idle events
		if prevEpoch != 0 && prevFilePath != "" {
			for e := prevEpoch + epoch.WindowSize; e < fileEpoch && e <= prevEpoch+epoch.IdleTimeout; e += epoch.WindowSize {
				if _, ok := events[e]; !ok {
					events[e] = make(map[string]int, 0)
				}
				events[e][prevFilePath]++
			}
		}
		prevEpoch = fileEpoch
		prevFilePath = filePath
	}

	// Add idle events for last event
	epochNow := epoch.MinuteNow()
	if prevEpoch != 0 && prevFilePath != "" {
		for e := prevEpoch + epoch.WindowSize; e < epochNow && e <= prevEpoch+epoch.IdleTimeout; e += epoch.WindowSize {
			if _, ok := events[e]; !ok {
				events[e] = make(map[string]int, 0)
			}
			events[e][prevFilePath]++
		}
	}

	if !dryRun {
		if err := removeFiles(filesToRemove); err != nil {
			return events, err
		}
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
