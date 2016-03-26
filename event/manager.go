package event

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"edgeg.io/gtm/epoch"
)

func Record(file string) error {
	rootPath, relFilePath, gtmPath, err := findPaths(file)
	if err != nil {
		return err
	}

	if err := writeEventFile(rootPath, relFilePath, gtmPath); err != nil {
		return err
	}

	return nil
}

func Process(gtmPath string, dryRun bool) (map[int64]map[string]int, error) {
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
