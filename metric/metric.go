package metric

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/scm"
)

func getFileID(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

// allocateTime calculates access time for each file within an epoch window
func allocateTime(ep int64, metricMap map[string]metricFile, eventMap map[string]int) error {
	total := 0
	for file := range eventMap {
		total += eventMap[file]
	}

	lastFile := ""
	timeAllocated := 0
	for file := range eventMap {
		t := int(float64(eventMap[file]) / float64(total) * float64(epoch.WindowSize))
		fileID := getFileID(file)

		var (
			mf  metricFile
			ok  bool
			err error
		)
		mf, ok = metricMap[fileID]
		if !ok {
			mf, err = newMetricFile(file, 0, true, map[int64]int{})
			if err != nil {
				return err
			}
		}
		mf.addTime(ep, t)

		//NOTE - Go has some gotchas when it comes to structs contained within maps
		// a copy is returned and not the reference to the struct
		// https://groups.google.com/forum/#!topic/golang-nuts/4_pabWnsMp0
		// assigning the new & updated metricFile instance to the map
		metricMap[fileID] = mf

		timeAllocated += t
		lastFile = file
	}
	// let's make sure all of the EpochWindowSize seconds are allocated
	// we put the remaining on the last file
	if lastFile != "" && timeAllocated < epoch.WindowSize {
		mf := metricMap[getFileID(lastFile)]
		mf.addTime(ep, epoch.WindowSize-timeAllocated)
	}
	return nil
}

type metricFile struct {
	// Updated signifies if we need to save the metric file
	Updated    bool
	SourceFile string
	TimeSpent  int
	GitTracked bool
	Timeline   map[int64]int
}

func (m *metricFile) addTime(ep int64, t int) {
	m.Updated = true
	m.TimeSpent += t
	m.Timeline[ep] += t
}

type metricFileByTime []metricFile

func (a metricFileByTime) Len() int           { return len(a) }
func (a metricFileByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a metricFileByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }

func newMetricFile(f string, t int, updated bool, timeline map[int64]int) (metricFile, error) {
	tracked, err := scm.GitTracked(f)
	if err != nil {
		return metricFile{}, err
	}

	return metricFile{SourceFile: f, TimeSpent: t, Updated: updated, GitTracked: tracked, Timeline: timeline}, nil
}

func marshalMetricFile(mf metricFile) []byte {
	s := fmt.Sprintf("%s:%d", mf.SourceFile, mf.TimeSpent)
	for e, t := range mf.Timeline {
		s += fmt.Sprintf(",%d:%d", e, t)
	}
	return []byte(s)
}

func unMarshalMetricFile(b []byte, filePath string) (metricFile, error) {
	var (
		fileName       string
		totalTimeSpent int
		err            error
	)

	timeline := map[int64]int{}
	parts := strings.Split(string(b), ",")

	for i := 0; i < len(parts); i++ {
		subparts := strings.Split(parts[i], ":")
		if len(subparts) != 2 {
			return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid format", filePath)
		}
		if i == 0 {
			fileName = subparts[0]
			totalTimeSpent, err = strconv.Atoi(subparts[1])
			if err != nil {
				return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid time, %s", filePath, err)
			}
			continue
		}
		ep, err := strconv.ParseInt(subparts[0], 10, 64)
		if err != nil {
			return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid epoch, %s", filePath, err)
		}
		timeSpent, err := strconv.Atoi(subparts[1])
		if err != nil {
			return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid time,  %s", filePath, err)
		}
		timeline[ep] += timeSpent
	}

	mf, err := newMetricFile(fileName, totalTimeSpent, false, timeline)
	if err != nil {
		return metricFile{}, err
	}

	return mf, nil
}

func loadMetrics(gtmPath string) (map[string]metricFile, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	metrics := map[string]metricFile{}
	for _, file := range files {

		if !strings.HasSuffix(file.Name(), ".metric") {
			continue
		}

		metricFilePath := filepath.Join(gtmPath, file.Name())

		metricFile, err := readMetricFile(metricFilePath)
		if err != nil {
			// TODO: purge bad metric files and log error
			continue
		}
		metrics[strings.Replace(file.Name(), ".metric", "", 1)] = metricFile
	}

	return metrics, nil
}

func saveMetrics(gtmPath string, metricMap map[string]metricFile, commitMap map[string]metricFile) error {
	for fileID, mf := range metricMap {
		_, inCommitMap := commitMap[fileID]

		if mf.Updated && !inCommitMap {
			// source file has updated time and is not in the commit
			if err := writeMetricFile(gtmPath, mf); err != nil {
				return err
			}
		}
		modified, err := scm.GitModified(mf.SourceFile)
		if err != nil {
			return err
		}
		if inCommitMap || (!inCommitMap && mf.GitTracked && !modified) {
			// source file is in commit or it's git tracked and not modified
			if err := removeMetricFile(gtmPath, fileID); err != nil {
				return err
			}
		}
	}
	return nil
}

func readMetricFile(filePath string) (metricFile, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return metricFile{}, err
	}

	return unMarshalMetricFile(b, filePath)
}

func writeMetricFile(gtmPath string, mf metricFile) error {
	if err := ioutil.WriteFile(
		filepath.Join(gtmPath, fmt.Sprintf("%s.metric", getFileID(mf.SourceFile))),
		marshalMetricFile(mf), 0644); err != nil {
		return err
	}

	return nil
}

func removeMetricFile(gtmPath, fileID string) error {
	fp := filepath.Join(gtmPath, fmt.Sprintf("%s.metric", fileID))
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(fp); err != nil {
		return err
	}

	return nil
}
