package metric

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"edgeg.io/gtm/scm"
)

type FileLog struct {
	FileMetric
	status string
}

type TimeLog struct {
	Version string
	Files   []FileLog
}

func marshalTimeLog(tl TimeLog) string {
	s := fmt.Sprintf("[ver:%s,total:%d]\n", "1", tl.Total())
	for _, fl := range tl.Files {
		s += fmt.Sprintf("%s,%s\n", marshalMetricFile(fl.FileMetric), fl.status)
	}
	return s
}

func unMarshalTimeLog(s string) (TimeLog, error) {
	var version string
	reHeader := regexp.MustCompile(`\[ver:\d+,total:\d+]`)
	reVersion := regexp.MustCompile(`\d+`)

	lines := strings.Split(s, "\n")
	tl := TimeLog{}
	for i := 0; i < len(lines); i++ {
		if reHeader.MatchString(lines[i]) {
			version = reVersion.FindString(lines[i])
			continue
		}
		switch version {
		case "1":
		default:
			return tl, fmt.Errorf("Unable to unmarshal time logged, unknown version %s", version)
		}
	}
	return tl, nil
}

func (t TimeLog) Total() int {
	total := 0
	for _, mf := range t.Files {
		total += mf.TimeSpent
	}
	return total
}

type FileLogByTime []FileLog

func (a FileLogByTime) Len() int      { return len(a) }
func (a FileLogByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a FileLogByTime) Less(i, j int) bool {
	return a[i].FileMetric.TimeSpent < a[j].FileMetric.TimeSpent
}

func NewTimeLog(metricMap map[string]FileMetric, commitMap map[string]FileMetric) (TimeLog, error) {
	fls := []FileLog{}
	for _, mf := range commitMap {
		mf.Downsample()
		fls = append(fls, FileLog{FileMetric: mf, status: "m"})
	}

	for fileID, mf := range metricMap {
		if _, ok := commitMap[fileID]; !ok {
			// looking at only files not in commit
			modified, err := scm.GitModified(mf.SourceFile)
			if err != nil {
				return TimeLog{}, err
			}
			if mf.GitTracked && !modified {
				// source file is tracked by git and is not modified
				mf.Downsample()
				fls = append(fls, FileLog{FileMetric: mf, status: "r"})
			}
		}
	}
	sort.Sort(sort.Reverse(FileLogByTime(fls)))
	return TimeLog{Version: "1", Files: fls}, nil
}
