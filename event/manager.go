// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/epoch"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
)

// Record creates an event for a source
func Record(file string) error {
	sourcePath, gtmPath, err := pathFromSource(file)
	if err != nil {
		return err
	}

	if err := writeEventFile(sourcePath, gtmPath); err != nil {
		return err
	}

	if err := project.SetActive(gtmPath); err != nil {
		return err
	}

	return nil
}

// Process scans the gtmPath for event files and processes them.
// If interim is true, event files are not purged.
func Process(gtmPath string, interim bool) (map[int64]map[string]int, error) {
	defer util.TimeTrack(time.Now(), "event.Process")

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

		sourcePath, err := readEventFile(eventFilePath)
		if err != nil {
			os.Remove(eventFilePath)
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][sourcePath]++

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
		prevFilePath = sourcePath
	}

	if !interim {
		if err := removeFiles(filesToRemove); err != nil {
			return events, err
		}
	}

	return events, nil
}
