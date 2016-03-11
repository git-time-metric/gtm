package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"edgeg.io/gtm/cfg"
	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/event"
	"github.com/dickeyxxx/golock"
)

type Metrics struct {
	Files map[string]int `json:"files"`
}

func (m *Metrics) AddTime(f string, t int) {
	m.Files[f] += t
}

func ProcessEvents() error {
	epochMarker := epoch.MinutePast()

	_, gtmPath, err := cfg.Paths()
	if err != nil {
		return err
	}

	lockFile := filepath.Join(gtmPath, "gtm.lock")
	if err := golock.Lock(lockFile); err != nil {
		return err
	}
	defer golock.Unlock(lockFile)

	eventMap, err := event.Sweep(epochMarker, gtmPath)
	if err != nil {
		return err
	}

	metrics, err := read(gtmPath)
	if err != nil {
		return err
	}

	for epoch := range eventMap {
		fileMap := make(map[string]int)
		total := 0
		for file := range eventMap[epoch] {
			total += eventMap[epoch][file]
			fileMap[file] += eventMap[epoch][file]
		}
		allocateTime(metrics, fileMap, total)
	}

	if err := save(gtmPath, metrics); err != nil {
		return err
	}

	return nil
}

func allocateTime(metrics *Metrics, fileMap map[string]int, total int) {
	var timeAllocated int
	var lastFile string
	for file := range fileMap {
		dur := int(float64(fileMap[file]) / float64(total) * float64(epoch.WindowSize))
		metrics.AddTime(file, dur)
		timeAllocated += dur
		lastFile = file
	}
	//let's make sure all of the EpochWindowSize seconds is allocated
	//we put the remaining on the last list of events
	if lastFile != "" && timeAllocated < epoch.WindowSize {
		metrics.AddTime(lastFile, epoch.WindowSize-timeAllocated)
	}
}

func read(gtmPath string) (*Metrics, error) {
	ms := Metrics{}
	fp := filepath.Join(gtmPath, "metrics.json")
	if b, err := ioutil.ReadFile(string(fp)); err == nil {
		if err := json.Unmarshal(b, &ms); err != nil {
			return &ms, fmt.Errorf("Reading metrics file failed with error %s", err)
		}
	}
	return &ms, nil
}

func save(gtmPath string, metrics *Metrics) error {
	if b, err := json.Marshal(metrics); err != nil {
		return fmt.Errorf("Save metrics to file failed with error %s", err)
	} else {
		fp := filepath.Join(gtmPath, "metrics.json")
		if err := ioutil.WriteFile(fp, b, 0644); err != nil {
			return fmt.Errorf("Saving metrics to file failed with error %s", err)
		}
	}
	return nil
}
