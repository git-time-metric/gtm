package metric

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"edgeg.io/gtm/cfg"
	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/event"
	"github.com/dickeyxxx/golock"
)

func Process() error {
	_, gtmPath, err := cfg.Paths()
	if err != nil {
		return err
	}

	lockFile := filepath.Join(gtmPath, "gtm.lock")
	if err := golock.Lock(lockFile); err != nil {
		return err
	}
	defer golock.Unlock(lockFile)

	eventMap, err := event.Sweep(gtmPath)
	if err != nil {
		return err
	}

	metricMap, err := load(gtmPath)
	if err != nil {
		return err
	}

	for epoch := range eventMap {
		allocateTime(metricMap, eventMap[epoch])
	}

	fmt.Printf("%+v\n", eventMap)
	fmt.Printf("%+v\n", metricMap)

	return nil
}

func fileID(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

func allocateTime(metricMap map[string]int, fileMap map[string]int) {
	total := 0
	for file := range fileMap {
		total += fileMap[file]
	}

	lastFile := ""
	timeAllocated := 0
	for file := range fileMap {
		dur := int(float64(fileMap[file]) / float64(total) * float64(epoch.WindowSize))
		metricMap[fileID(file)] += dur
		timeAllocated += dur
		lastFile = file
	}
	//let's make sure all of the EpochWindowSize seconds is allocated
	//we put the remaining on the last list of events
	if lastFile != "" && timeAllocated < epoch.WindowSize {
		metricMap[fileID(lastFile)] += epoch.WindowSize - timeAllocated
	}
}

func load(gtmPath string) (map[string]int, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	metrics := map[string]int{}
	for _, file := range files {

		if !strings.HasSuffix(file.Name(), ".metric") {
			continue
		}

		metricFilePath := filepath.Join(gtmPath, file.Name())

		t, err := read(metricFilePath)
		if err != nil {
			continue
		}
		metrics[file.Name()] = t
	}

	return metrics, nil
}

func read(filePath string) (int, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

func write(gtmPath, fileID string, t int) error {
	if err := ioutil.WriteFile(
		filepath.Join(
			gtmPath,
			fmt.Sprintf("%s.metric", fileID)),
		[]byte(strconv.Itoa(t)),
		0644); err != nil {
		return err
	}

	return nil
}
