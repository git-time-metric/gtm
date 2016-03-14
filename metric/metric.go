package metric

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/epoch"
	"edgeg.io/gtm/event"
	"edgeg.io/gtm/scm"
	"github.com/dickeyxxx/golock"
)

func Process(dryRun bool) error {
	_, gtmPath, err := env.Paths()
	if err != nil {
		return err
	}

	lockFile := filepath.Join(gtmPath, "gtm.lock")
	if err := golock.Lock(lockFile); err != nil {
		return err
	}
	defer golock.Unlock(lockFile)

	epochEventMap, err := event.Sweep(gtmPath, dryRun)
	if err != nil {
		return err
	}

	metricMap, err := loadMetrics(gtmPath)
	if err != nil {
		return err
	}

	for epoch := range epochEventMap {
		err := allocateTime(metricMap, epochEventMap[epoch])
		if err != nil {
			return err
		}
	}

	m, err := scm.GitCommitMsg()
	if err != nil {
		return err
	}
	_, _, commitFiles := scm.GitParseMessage(m)

	commitMap := map[string]metricFile{}
	if !dryRun {
		//for only files in the last commit
		for _, f := range commitFiles {
			fileID := getFileID(f)
			if _, ok := metricMap[fileID]; !ok {
				continue
			}
			commitMap[fileID] = metricMap[fileID]
		}
	}

	writeNote(gtmPath, metricMap, commitMap, dryRun)
	saveMetrics(gtmPath, metricMap, commitMap, dryRun)

	log.Printf("epochEventMap -> %+v", epochEventMap)
	log.Printf("metricMap -> %+v", metricMap)
	log.Printf("commitMap -> %+v", commitMap)
	log.Printf("dryRun -> %+v", dryRun)

	return nil
}

func getFileID(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

func allocateTime(metricMap map[string]metricFile, eventMap map[string]int) error {
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
			mf, err = newMetricFile(file, 0, true)
			if err != nil {
				return err
			}
		}
		mf.AddTime(t)

		//NOTE - Go has some gotchas when it comes to structs contained within maps
		// a copy is returned and not the reference to the struct
		// https://groups.google.com/forum/#!topic/golang-nuts/4_pabWnsMp0
		// assigning the new & updated metricFile instance to the map
		metricMap[fileID] = mf

		timeAllocated += t
		lastFile = file
	}
	// let's make sure all of the EpochWindowSize seconds are allocated
	// we put the remaining on the last list of events
	if lastFile != "" && timeAllocated < epoch.WindowSize {
		mf := metricMap[getFileID(lastFile)]
		mf.AddTime(epoch.WindowSize - timeAllocated)
	}
	return nil
}

type metricFile struct {
	Updated bool // Updated signifies if we need to save metric file
	GitFile string
	Time    int
}

func (m *metricFile) AddTime(t int) {
	m.Updated = true
	m.Time += t
}

func (m *metricFile) GitTracked() bool {
	tracked, err := scm.GitTracked(m.GitFile)
	if err != nil {
		// for ease of use, let's squash errors
		log.Printf("%s", err)
		return false
	}

	return tracked
}

func (m *metricFile) GitModified() bool {
	modified, err := scm.GitModified(m.GitFile)
	if err != nil {
		// for ease of use, let's squash errors
		log.Printf("%s", err)
		return false
	}

	return modified
}

func newMetricFile(f string, t int, updated bool) (metricFile, error) {
	return metricFile{GitFile: f, Time: t, Updated: updated}, nil
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
			continue
		}
		metrics[strings.Replace(file.Name(), ".metric", "", 1)] = metricFile
	}

	return metrics, nil
}

func saveMetrics(gtmPath string, metricMap map[string]metricFile, commitMap map[string]metricFile, dryRun bool) error {
	if !dryRun {
		for fileID, mf := range metricMap {
			_, inCommit := commitMap[fileID]
			if mf.Updated && !inCommit {
				writeMetricFile(gtmPath, mf)
			}
			// remove files in commit or
			// remove git tracked and not modified files not in commit
			if inCommit || (!inCommit && mf.GitTracked() && !mf.GitModified()) {
				removeMetricFile(gtmPath, fileID)
			}
		}
	}
	return nil
}

func readMetricFile(filePath string) (metricFile, error) {
	log.Printf("readMetricFile -> %+v", filePath)
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return metricFile{}, err
	}

	parts := strings.Split(string(b), ",")
	if len(parts) < 2 {
		return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid contents", filePath)
	}

	t, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return metricFile{}, fmt.Errorf("Unable to parse metric file %s, invalid time -> %s", err)
	}

	mf, err := newMetricFile(parts[0], t, false)
	if err != nil {
		return metricFile{}, err
	}

	return mf, nil
}

func writeMetricFile(gtmPath string, mf metricFile) error {
	log.Printf("writeMetricFile -> %+v", mf)
	if err := ioutil.WriteFile(
		filepath.Join(gtmPath, fmt.Sprintf("%s.metric", getFileID(mf.GitFile))),
		[]byte(fmt.Sprintf("%s,%d", mf.GitFile, mf.Time)), 0644); err != nil {
		return err
	}

	return nil
}

func removeMetricFile(gtmPath, fileID string) error {
	log.Printf("removeMetricFile -> %+v", fileID)
	if err := os.Remove(
		filepath.Join(
			gtmPath, fmt.Sprintf("%s.metric", fileID))); err != nil {
		return err
	}

	return nil
}

type metricFilePair struct {
	Key   string
	Value metricFile
}

type metricFileList []metricFilePair

func newMetricFileList(m map[string]metricFile) metricFileList {
	mfs := make(metricFileList, len(m))
	i := 0
	for k, v := range m {
		mfs[i] = metricFilePair{k, v}
		i++
	}
	return mfs
}

func (p metricFileList) Len() int           { return len(p) }
func (p metricFileList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p metricFileList) Less(i, j int) bool { return p[i].Value.Time < p[j].Value.Time }

func writeNote(gtmPath string, metricMap map[string]metricFile, commitMap map[string]metricFile, dryRun bool) error {
	if dryRun {
		commitMap = map[string]metricFile{}
		for fileID, mf := range metricMap {
			//include modified and git tracked files in commit map
			if mf.GitTracked() && mf.GitModified() {
				commitMap[fileID] = mf
			}
		}
	}

	var (
		total int
		note  string
	)

	commitList := newMetricFileList(commitMap)
	sort.Sort(sort.Reverse(commitList))
	for _, mf := range commitList {
		total += mf.Value.Time
		note += fmt.Sprintf("%s: %d [m]\n", mf.Value.GitFile, mf.Value.Time)
	}

	metricList := newMetricFileList(metricMap)
	sort.Sort(metricList)
	for _, mf := range metricList {
		// include git tracked and not modified files not in commit
		if _, ok := commitMap[mf.Key]; !ok && mf.Value.GitTracked() && !mf.Value.GitModified() {
			total += mf.Value.Time
			note += fmt.Sprintf("%s: %d [r]\n", mf.Value.GitFile, mf.Value.Time)
		}
	}
	note = fmt.Sprintf("total: %d\n", total) + note

	if dryRun {
		fmt.Print(note)
	} else {
		err := scm.GitAddNote(note)
		if err != nil {
			return err
		}
	}

	return nil
}
