package note

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"edgeg.io/gtm/util"
)

type CommitNote struct {
	Files []FileDetail
}

func (n CommitNote) Total() int {
	total := 0
	for _, fm := range n.Files {
		total += fm.TimeSpent
	}
	return total
}

func Marshal(n CommitNote) string {
	//TODO use a text template here instead
	s := fmt.Sprintf("[ver:%s,total:%d]\n", "1", n.Total())
	for _, fl := range n.Files {
		s += fmt.Sprintf("%s:%d,", fl.SourceFile, fl.TimeSpent)
		for _, e := range fl.SortEpochs() {
			s += fmt.Sprintf("%d:%d,", e, fl.Timeline[e])
		}
		s += fmt.Sprintf("%s\n", fl.Status)
	}
	return s
}

func UnMarshal(s string) (CommitNote, error) {
	var (
		version string
		files   = []FileDetail{}
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
				return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, header format invalid, %s", lines[lineIdx])
			}
		case version == "1":
			fieldGroups := strings.Split(lines[lineIdx], ",")
			if len(fieldGroups) < 3 {
				return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", lines[lineIdx])
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
					// file name and total, filename:total
					filePath = fieldVals[0]
					t, err := strconv.Atoi(fieldVals[1])
					if err != nil {
						return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					fileTotal = t
				case groupIdx == len(fieldGroups)-1 && len(fieldVals) == 1:
					// file status of m or r
					fileStatus = fieldVals[0]
				case len(fieldVals) == 2:
					// epoch timeline, epoch:total
					e, err := strconv.ParseInt(fieldVals[0], 10, 64)
					if err != nil {
						return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					t, err := strconv.Atoi(fieldVals[1])
					if err != nil {
						return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, format invalid, %s", err)
					}
					fileTimeline[e] = t
				default:
					// error
					return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, format invalid")
				}
			}

			// check for existing file path and merge if found
			// for example, this can happen when rewriting commits with git commit --amend
			found := false
			for idx, _ := range files {
				if files[idx].SourceFile == filePath {
					for epoch, secs := range fileTimeline {
						files[idx].TimeSpent += secs
						files[idx].Timeline[epoch] += secs
					}
					// only change file status if modified or deleted
					if fileStatus == "m" || fileStatus == "d" {
						files[idx].Status = fileStatus
					}
					found = true
					break
				}
			}

			if !found {
				files = append(files,
					FileDetail{
						SourceFile: filePath,
						TimeSpent:  fileTotal,
						Timeline:   fileTimeline,
						Status:     fileStatus})
			}

		default:
			return CommitNote{}, fmt.Errorf("Unable to unmarshal time logged, unknown version %s", version)
		}
	}
	return CommitNote{Files: files}, nil
}

type FileDetail struct {
	SourceFile string
	TimeSpent  int
	Timeline   map[int64]int
	Status     string
}

func (f *FileDetail) SortEpochs() []int64 {
	keys := []int64{}
	for k := range f.Timeline {
		keys = append(keys, k)
	}
	sort.Sort(util.ByInt64(keys))
	return keys
}

type FileByTime []FileDetail

func (a FileByTime) Len() int           { return len(a) }
func (a FileByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a FileByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }
