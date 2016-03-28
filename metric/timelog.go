package metric

import (
	"sort"

	"edgeg.io/gtm/scm"
)

type timeLogged struct {
	Files    []metricFile
	modified map[string]bool
}

func (t timeLogged) Total() int {
	total := 0
	for _, mf := range t.Files {
		total += mf.TimeSpent
	}
	return total
}

func (t timeLogged) FileStatus(f string) string {
	if t.modified[f] {
		return "m"
	}
	return "r"
}

func newTimeLogged(metricMap map[string]metricFile, commitMap map[string]metricFile) (timeLogged, error) {
	mfs := []metricFile{}
	modifiedMap := map[string]bool{}
	for _, mf := range commitMap {
		mfs = append(mfs, mf)
		// any files in the commit are tagged as modified
		// used for reporting status of file - modified or readonly
		modifiedMap[mf.SourceFile] = true
	}

	for fileID, mf := range metricMap {
		if _, ok := commitMap[fileID]; !ok {
			// looking at only files not in commit
			modified, err := scm.GitModified(mf.SourceFile)
			if err != nil {
				return timeLogged{}, err
			}
			if mf.GitTracked && !modified {
				// source file is tracked by git and is not modified
				mfs = append(mfs, mf)
			}
		}
	}
	sort.Sort(sort.Reverse(metricFileByTime(mfs)))
	return timeLogged{Files: mfs, modified: modifiedMap}, nil
}
