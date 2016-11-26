// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package metric

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/git-time-metric/gtm/epoch"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

// getFileID returns the SHA1 checksum for filePath
func getFileID(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filepath.ToSlash(filePath))))
}

// allocateTime calculates access time for each file within an epoch window
func allocateTime(ep int64, metricMap map[string]FileMetric, eventMap map[string]int) error {
	total := 0
	for file := range eventMap {
		total += eventMap[file]
	}

	lastFileID := ""
	timeAllocated := 0
	for file := range eventMap {
		t := int(float64(eventMap[file]) / float64(total) * float64(epoch.WindowSize))
		fileID := getFileID(file)

		var (
			fm  FileMetric
			ok  bool
			err error
		)
		fm, ok = metricMap[fileID]
		if !ok {
			fm, err = newFileMetric(file, 0, true, map[int64]int{})
			if err != nil {
				return err
			}
		}
		fm.AddTimeSpent(ep, t)

		//NOTE: Go has some gotchas when it comes to structs contained within maps
		// a copy is returned and not the reference to the struct
		// https://groups.google.com/forum/#!topic/golang-nuts/4_pabWnsMp0
		// assigning the new & updated metricFile instance to the map
		metricMap[fileID] = fm

		timeAllocated += t
		lastFileID = fileID
	}
	// let's make sure all of the EpochWindowSize seconds are allocated
	// we put the remaining on the last file
	if lastFileID != "" && timeAllocated < epoch.WindowSize {
		fm := metricMap[lastFileID]
		fm.AddTimeSpent(ep, epoch.WindowSize-timeAllocated)
		metricMap[lastFileID] = fm
	}
	return nil
}

// FileMetric contains the source file and it's time metrics
type FileMetric struct {
	Updated    bool // Updated signifies if we need to save the metric file
	SourceFile string
	TimeSpent  int
	Timeline   map[int64]int
}

// AddTimeSpent accumulates time spent for a source file
func (f *FileMetric) AddTimeSpent(ep int64, t int) {
	f.Updated = true
	f.TimeSpent += t
	f.Timeline[ep] += t
}

// Downsample return timeline by hour
func (f *FileMetric) Downsample() {
	byHour := map[int64]int{}
	for ep, t := range f.Timeline {
		byHour[ep/3600*3600] += t
	}
	f.Timeline = byHour
}

// SortEpochs returns sorted timeline epochs
func (f *FileMetric) SortEpochs() []int64 {
	keys := []int64{}
	for k := range f.Timeline {
		keys = append(keys, k)
	}
	sort.Sort(util.ByInt64(keys))
	return keys
}

// FileMetricByTime is an array of FileMetrics
type FileMetricByTime []FileMetric

func (a FileMetricByTime) Len() int           { return len(a) }
func (a FileMetricByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a FileMetricByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }

func newFileMetric(f string, t int, updated bool, timeline map[int64]int) (FileMetric, error) {
	return FileMetric{SourceFile: f, TimeSpent: t, Updated: updated, Timeline: timeline}, nil
}

// marshalFileMetric converts FileMetric struct to a byte array
func marshalFileMetric(fm FileMetric) []byte {
	s := fmt.Sprintf("%s:%d", fm.SourceFile, fm.TimeSpent)
	for _, e := range fm.SortEpochs() {
		s += fmt.Sprintf(",%d:%d", e, fm.Timeline[e])
	}
	return []byte(s)
}

// unMarshalFileMetric converts a byte array to a FileMetric struct
func unMarshalFileMetric(b []byte, filePath string) (FileMetric, error) {
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
			return FileMetric{}, fmt.Errorf("Unable to parse metric file %s, invalid format", filePath)
		}
		if i == 0 {
			fileName = subparts[0]
			totalTimeSpent, err = strconv.Atoi(subparts[1])
			if err != nil {
				return FileMetric{}, fmt.Errorf("Unable to parse metric file %s, invalid time, %s", filePath, err)
			}
			continue
		}
		ep, err := strconv.ParseInt(subparts[0], 10, 64)
		if err != nil {
			return FileMetric{}, fmt.Errorf("Unable to parse metric file %s, invalid epoch, %s", filePath, err)
		}
		timeSpent, err := strconv.Atoi(subparts[1])
		if err != nil {
			return FileMetric{}, fmt.Errorf("Unable to parse metric file %s, invalid time,  %s", filePath, err)
		}
		timeline[ep] += timeSpent
	}

	fm, err := newFileMetric(fileName, totalTimeSpent, false, timeline)
	if err != nil {
		return FileMetric{}, err
	}

	return fm, nil
}

// loadMetrics scans the gtmPath for metric files and loads them
func loadMetrics(gtmPath string) (map[string]FileMetric, error) {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return nil, err
	}

	metrics := map[string]FileMetric{}
	for _, file := range files {

		if !strings.HasSuffix(file.Name(), ".metric") {
			continue
		}

		metricFilePath := filepath.Join(gtmPath, file.Name())

		metricFile, err := readMetricFile(metricFilePath)
		if err != nil {
			os.Remove(metricFilePath)
			continue
		}

		metrics[strings.Replace(file.Name(), ".metric", "", 1)] = metricFile
	}

	return metrics, nil
}

// saveAndPurgeMetrics deletes metric files that are in the commit and save any that are not
func saveAndPurgeMetrics(
	gtmPath string,
	metricMap map[string]FileMetric,
	commitMap map[string]FileMetric,
	readonlyMap map[string]FileMetric) error {

	for fileID, fm := range metricMap {
		_, inCommitMap := commitMap[fileID]
		_, inReadonlyMap := readonlyMap[fileID]

		//Save metric files that are updated and not in commit or readonly maps
		if fm.Updated && !inCommitMap && !inReadonlyMap {
			if err := writeMetricFile(gtmPath, fm); err != nil {
				return err
			}
		}

		//Purge metric files that are in the commit and readonly maps
		if inCommitMap || inReadonlyMap {
			if err := removeMetricFile(gtmPath, fileID); err != nil {
				return err
			}
		}
	}
	return nil
}

// readMetric reads and returns the unmarshalled metric file
func readMetricFile(filePath string) (FileMetric, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return FileMetric{}, err
	}

	return unMarshalFileMetric(b, filePath)
}

// writeMetricFile persists metric file to disk
func writeMetricFile(gtmPath string, fm FileMetric) error {
	if err := ioutil.WriteFile(
		filepath.Join(gtmPath, fmt.Sprintf("%s.metric", getFileID(fm.SourceFile))),
		marshalFileMetric(fm), 0644); err != nil {
		return err
	}

	return nil
}

// removeMetricFile deletes a metric file with fileID
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

// buildCommitMaps creates the write and read-only commit maps.
// Files that are in the head commit are added to write commit map.
// Files that are are not in the commit map and are readonly are added to the read-only commit map.
func buildCommitMaps(metricMap map[string]FileMetric) (map[string]FileMetric, map[string]FileMetric, error) {

	commitMap := map[string]FileMetric{}
	readonlyMap := map[string]FileMetric{}

	commit, err := scm.HeadCommit()
	if err != nil {
		return commitMap, readonlyMap, err
	}

	for _, f := range commit.Stats.Files {
		fileID := getFileID(f)
		if _, ok := metricMap[fileID]; !ok {
			continue
		}
		commitMap[fileID] = metricMap[fileID]
	}

	for fileID, fm := range metricMap {
		// Look at files not in commit map
		if _, ok := commitMap[fileID]; !ok {
			status, err := scm.NewStatus()
			if err != nil {
				return commitMap, readonlyMap, err
			}

			if !status.IsModified(fm.SourceFile, false) {
				readonlyMap[fileID] = fm
			}
		}
	}

	return commitMap, readonlyMap, nil
}

// buildCommitNote creates a CommitNote for files in the commit and readonly maps in git repo at rootPath
func buildCommitNote(
	rootPath string,
	commitMap map[string]FileMetric,
	readonlyMap map[string]FileMetric) (note.CommitNote, error) {

	flsModified := []note.FileDetail{}

	for _, fm := range commitMap {
		fm.Downsample()
		status := "m"
		if _, err := os.Stat(filepath.Join(rootPath, fm.SourceFile)); os.IsNotExist(err) {
			status = "d"
		}
		flsModified = append(
			flsModified,
			note.FileDetail{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: status})
	}

	flsReadonly := []note.FileDetail{}
	for _, fm := range readonlyMap {
		fm.Downsample()
		status := "r"
		if _, err := os.Stat(filepath.Join(rootPath, fm.SourceFile)); os.IsNotExist(err) {
			status = "d"
		}
		flsReadonly = append(
			flsReadonly,
			note.FileDetail{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: status})
	}
	fls := append(flsModified, flsReadonly...)
	sort.Sort(sort.Reverse(note.FileByTime(fls)))

	return note.CommitNote{Files: fls}, nil
}

// buildInterimCommitMaps creates the write and read-only commit maps
// Write and read-only files maps are built based on an algorithm and not a git commit
func buildInterimCommitMaps(metricMap map[string]FileMetric, projPath ...string) (map[string]FileMetric, map[string]FileMetric, error) {
	commitMap := map[string]FileMetric{}
	readonlyMap := map[string]FileMetric{}

	status, err := scm.NewStatus(projPath...)
	if err != nil {
		return commitMap, readonlyMap, err
	}

	for fileID, fm := range metricMap {
		if status.HasStaged() {
			if status.IsModified(fm.SourceFile, true) {
				commitMap[fileID] = fm
			} else {
				// when in staging, include any files in working that are not modified
				if !status.IsModified(fm.SourceFile, false) {
					readonlyMap[fileID] = fm
				}
			}
		} else {
			if status.IsModified(fm.SourceFile, false) {
				commitMap[fileID] = fm
			} else {
				readonlyMap[fileID] = fm
			}
		}
	}

	return commitMap, readonlyMap, nil
}
