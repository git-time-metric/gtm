package metric

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"edgeg.io/gtm/scm"
)

type FileLog struct {
	SourceFile string
	TimeSpent  int
	Timeline   map[int64]int
	Status     string
}

func (f *FileLog) SortEpochs() []int64 {
	keys := []int64{}
	for k := range f.Timeline {
		keys = append(keys, k)
	}
	sort.Sort(ByEpoch(keys))
	return keys
}

func NewFileLog(filePath string, total int, timeline map[int64]int, status string) (FileLog, error) {
	return FileLog{SourceFile: filePath, TimeSpent: total, Timeline: timeline, Status: status}, nil
}

type TimeLog struct {
	Files []FileLog
}

func marshalTimeLog(tl TimeLog) string {
	s := fmt.Sprintf("[ver:%s,total:%d]\n", "1", tl.Total())
	for _, fl := range tl.Files {
		s += fmt.Sprintf("%s:%d,", fl.SourceFile, fl.TimeSpent)
		for _, e := range fl.SortEpochs() {
			s += fmt.Sprintf("%d:%d,", e, fl.Timeline[e])
		}
		s += fmt.Sprintf("%s", fl.Status)
	}
	return s
}

func unMarshalTimeLog(s string) (TimeLog, error) {
	var (
		version  string
		fileLogs = []FileLog{}
	)

	reHeader := regexp.MustCompile(`\[ver:\d+,total:\d+]`)
	reHeaderVals := regexp.MustCompile(`\d+`)

	lines := strings.Split(s, "\n")
	for lineIdx := 0; lineIdx < len(lines); lineIdx++ {
		switch {
		case strings.TrimSpace(lines[lineIdx]) == "":
			version = ""
		case reHeader.MatchString(lines[lineIdx]):
			if matches := reHeaderVals.FindAllString(lines[lineIdx], 2); matches != nil && len(matches) == 2 {
				version = matches[0]
			} else {
				return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, header format invalid, %s", lines[lineIdx])
			}
		case version == "1":
			fieldGroups := strings.Split(lines[lineIdx], ",")
			if len(fieldGroups) < 3 {
				return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", lines[lineIdx])
			}

			var (
				filePath     string
				fileTotal    int
				fileStatus   string
				fileTimeline = map[int64]int{}
			)

			for groupIdx := range fieldGroups {
				fieldVals := strings.Split(fieldGroups[groupIdx], ":")
				switch {
				case groupIdx == 0 && len(fieldVals) == 2:
					// file name and total
					filePath = fieldVals[0]
					t, err := strconv.Atoi(fieldVals[1])
					if err != nil {
						return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					fileTotal = t
				case groupIdx == len(fieldGroups)-1 && len(fieldVals) == 1:
					fileStatus = fieldVals[0]
				case len(fieldVals) == 2:
					e, err := strconv.ParseInt(fieldVals[0], 10, 64)
					if err != nil {
						return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					t, err := strconv.Atoi(fieldVals[1])
					if err != nil {
						return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					fileTimeline[e] = t
				default:
					// error
				}
			}
			fl, err := NewFileLog(filePath, fileTotal, fileTimeline, fileStatus)
			if err != nil {
				return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
			}
			fileLogs = append(fileLogs, fl)

		default:
			return TimeLog{}, fmt.Errorf("Unable to unmarshal time logged, unknown version %s", version)
		}
	}
	return TimeLog{Files: fileLogs}, nil
}

func (t TimeLog) Total() int {
	total := 0
	for _, fm := range t.Files {
		total += fm.TimeSpent
	}
	return total
}

type FileLogByTime []FileLog

func (a FileLogByTime) Len() int           { return len(a) }
func (a FileLogByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a FileLogByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }

func NewTimeLog(metricMap map[string]FileMetric, commitMap map[string]FileMetric) (TimeLog, error) {
	fls := []FileLog{}
	for _, fm := range commitMap {
		fm.Downsample()
		fls = append(fls, FileLog{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: "m"})
	}

	for fileID, fm := range metricMap {
		if _, ok := commitMap[fileID]; !ok {
			// looking at only files not in commit
			modified, err := scm.GitModified(fm.SourceFile)
			if err != nil {
				return TimeLog{}, err
			}
			if fm.GitTracked && !modified {
				// source file is tracked by git and is not modified
				fm.Downsample()
				fls = append(fls, FileLog{SourceFile: fm.SourceFile, TimeSpent: fm.TimeSpent, Timeline: fm.Timeline, Status: "r"})
			}
		}
	}
	sort.Sort(sort.Reverse(FileLogByTime(fls)))
	return TimeLog{Files: fls}, nil
}
