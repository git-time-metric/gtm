package event

import (
	"encoding/json"
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
	if !cfg.FileExist(file) {
		return cfg.ErrFileNotFound
	}

	filePath, err := cfg.FilePath(file)
	if err != nil {
		return err
	}

	rootPath, gtmPath, err := cfg.Paths(filePath)
	if err != nil {
		return err
	}

	relFilePath, err := cfg.RelativePath(file, rootPath)
	if err != nil {
		return err
	}

	if err := writeFile(relFilePath, gtmPath); err != nil {
		return err
	}

	Sweep(epoch.MinutePast(), gtmPath)

	return nil
}

type event struct {
	File string `json:"file"`
}

func writeFile(relFilePath, gtmPath string) error {
	j, err := json.Marshal(event{relFilePath})
	if err != nil {
		return err
	}

	eventFile := filepath.Join(
		gtmPath,
		fmt.Sprintf("event-%d-%s.json", epoch.MinuteNow(), uuid.NewV4().String()[:8]))

	if err := ioutil.WriteFile(eventFile, j, 0644); err != nil {
		return err
	}
	return nil
}

func Sweep(epochMarker int64, gtmPath string) (map[int64]map[string]int, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	events := make(map[int64]map[string]int, 0)
	removeFiles := []string{}
	for _, file := range files {

		if !strings.HasPrefix(file.Name(), "event-") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(gtmPath, file.Name())
		removeFiles = append(removeFiles, filePath)

		s := strings.SplitN(file.Name(), "-", 3)
		if len(s) < 3 {
			continue
		}

		fileEpoch, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			continue
		} else if fileEpoch > epochMarker {
			//not ready yet for processing
			//pop it off the removeFile list
			removeFiles = removeFiles[:len(removeFiles)]
			continue
		}

		b, err := ioutil.ReadFile(string(filePath))
		if err != nil {
			continue
		}

		var e event
		if err := json.Unmarshal(b, &e); err != nil {
			continue
		}

		if _, ok := events[fileEpoch]; !ok {
			events[fileEpoch] = make(map[string]int, 0)
		}
		events[fileEpoch][e.File] += 1
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
