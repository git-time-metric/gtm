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

	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
	"edgeg.io/gtm/util"
)

func getFileID(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

// allocateTime calculates access time for each file within an epoch window
func allocateTime(ep int64, metricMap map[string]FileMetric, eventMap map[string]int) error {
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

		//NOTE - Go has some gotchas when it comes to structs contained within maps
		// a copy is returned and not the reference to the struct
		// https://groups.google.com/forum/#!topic/golang-nuts/4_pabWnsMp0
		// assigning the new & updated metricFile instance to the map
		metricMap[fileID] = fm

		timeAllocated += t
		lastFile = file
	}
	// let's make sure all of the EpochWindowSize seconds are allocated
	// we put the remaining on the last file
	if lastFile != "" && timeAllocated < epoch.WindowSize {
		fm := metricMap[getFileID(lastFile)]
		fm.AddTimeSpent(ep, epoch.WindowSize-timeAllocated)
	}
	return nil
}

type FileMetric struct {
	// Updated signifies if we need to save the metric file
	Updated    bool
	SourceFile string
	TimeSpent  int
	GitTracked bool
	Timeline   map[int64]int
}

func (f *FileMetric) AddTimeSpent(ep int64, t int) {
	f.Updated = true
	f.TimeSpent += t
	f.Timeline[ep] += t
}

func (f *FileMetric) Downsample() {
	byHour := map[int64]int{}
	for ep, t := range f.Timeline {
		byHour[ep/3600*3600] += t
	}
	f.Timeline = byHour
}

func (f *FileMetric) SortEpochs() []int64 {
	keys := []int64{}
	for k := range f.Timeline {
		keys = append(keys, k)
	}
	sort.Sort(util.ByInt64(keys))
	return keys
}

type FileMetricByTime []FileMetric

func (a FileMetricByTime) Len() int           { return len(a) }
func (a FileMetricByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a FileMetricByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }

func newFileMetric(f string, t int, updated bool, timeline map[int64]int) (FileMetric, error) {
	tracked, err := scm.GitTracked(f)
	if err != nil {
		return FileMetric{}, err
	}

	return FileMetric{SourceFile: f, TimeSpent: t, Updated: updated, GitTracked: tracked, Timeline: timeline}, nil
}

func marshalFileMetric(fm FileMetric) []byte {
	s := fmt.Sprintf("%s:%d", fm.SourceFile, fm.TimeSpent)
	for _, e := range fm.SortEpochs() {
		s += fmt.Sprintf(",%d:%d", e, fm.Timeline[e])
	}
	return []byte(s)
}

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

func loadMetrics(rootPath, gtmPath string) (map[string]FileMetric, error) {
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
			project.Log(fmt.Sprintf("Removing corrupt metric file %s, %s", metricFilePath, err))
			if err := os.Remove(metricFilePath); err != nil {
				project.Log(fmt.Sprintf("Unable to delete corrupt metric file %s, %s", metricFilePath, err))
			}
			continue
		}

		// check for missing source file and delete metric file if it does not exist
		if fileInfo, err := os.Stat(filepath.Join(rootPath, metricFile.SourceFile)); os.IsNotExist(err) || fileInfo.IsDir() {
			if err := os.Remove(metricFilePath); err != nil {
				project.Log(fmt.Sprintf("\nError removing metric file %s, %s", metricFilePath, err))
			}
		}

		metrics[strings.Replace(file.Name(), ".metric", "", 1)] = metricFile
	}

	return metrics, nil
}

func saveAndPurgeMetrics(gtmPath string, metricMap map[string]FileMetric, commitMap map[string]FileMetric) error {
	for fileID, fm := range metricMap {
		_, inCommitMap := commitMap[fileID]

		if fm.Updated && !inCommitMap {
			// source file has updated time and is not in the commit
			if err := writeMetricFile(gtmPath, fm); err != nil {
				return err
			}
		}
		modified, err := scm.GitModified(fm.SourceFile, false)
		if err != nil {
			return err
		}
		if inCommitMap || (!inCommitMap && fm.GitTracked && !modified) {
			// source file is in commit or it's git tracked and not modified
			if err := removeMetricFile(gtmPath, fileID); err != nil {
				return err
			}
		}
	}
	return nil
}

func readMetricFile(filePath string) (FileMetric, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return FileMetric{}, err
	}

	return unMarshalFileMetric(b, filePath)
}

func writeMetricFile(gtmPath string, fm FileMetric) error {
	if err := ioutil.WriteFile(
		filepath.Join(gtmPath, fmt.Sprintf("%s.metric", getFileID(fm.SourceFile))),
		marshalFileMetric(fm), 0644); err != nil {
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

func buildCommitNote(metricMap map[string]FileMetric, commitMap map[string]FileMetric, gstate GitState) (note.CommitNote, error) {
	flsModified := []note.FileDetail{}

	if (gstate == Staging) && len(commitMap) == 0 {
		// When reporting on staging and no modified files, then don't report anything
		return note.CommitNote{}, nil
	}

	for _, fm := range commitMap {
		fm.Downsample()
		flsModified = append(flsModified, note.FileDetail{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: "m"})
	}

	var (
		staged bool
		err    error
	)
	staged, err = scm.GitHasStaged()
	if err != nil {
		return note.CommitNote{}, err
	}

	flsReadonly := []note.FileDetail{}
	for fileID, fm := range metricMap {
		// if staged file and looking at working, skip readonly
		// all the readonly files are allocated to staging
		if staged && gstate == Working {
			break
		}
		if _, ok := commitMap[fileID]; !ok {
			// looking at only files not in commit map
			modified, err := scm.GitModified(fm.SourceFile, false)
			if err != nil {
				return note.CommitNote{}, err
			}

			if fm.GitTracked && !modified {
				// source file is tracked by git and is not modified
				fm.Downsample()
				flsReadonly = append(flsReadonly, note.FileDetail{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: "r"})
			}
		}
	}
	fls := append(flsModified, flsReadonly...)
	sort.Sort(sort.Reverse(note.FileByTime(fls)))
	return note.CommitNote{Files: fls}, nil
}
