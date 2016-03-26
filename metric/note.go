package metric

import (
	"fmt"
	"sort"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/scm"
)

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
			if mf.gitTracked() && mf.gitModified() {
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
	sort.Sort(sort.Reverse(metricList))
	for _, mf := range metricList {
		// include git tracked and not modified files not in commit
		if _, ok := commitMap[mf.Key]; !ok && mf.Value.gitTracked() && !mf.Value.gitModified() {
			total += mf.Value.Time
			note += fmt.Sprintf("%s: %d [r]\n", mf.Value.GitFile, mf.Value.Time)
		}
	}
	note = fmt.Sprintf("\ntotal: %d\n", total) + note

	if dryRun {
		fmt.Print(note)
	} else {
		err := scm.GitAddNote(note, env.NoteNameSpace)
		if err != nil {
			return err
		}
	}

	return nil
}
