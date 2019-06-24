// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package note

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
)

// CommitNote contains the time metrics for a commit
type CommitNote struct {
	Files []FileDetail
}

// FilterOutTerminal filters out terminal time from commit note
func (n CommitNote) FilterOutTerminal() CommitNote {
	fds := []FileDetail{}
	for _, f := range n.Files {
		if !f.IsTerminal() {
			fds = append(fds, f)
		}
	}
	return CommitNote{Files: fds}
}

// FilterOutApp filters out app time from commit note
func (n CommitNote) FilterOutApp() CommitNote {
	fds := []FileDetail{}
	for _, f := range n.Files {
		if !f.IsApp() {
			fds = append(fds, f)
		}
	}
	return CommitNote{Files: fds}
}

// Total returns the total time for a commit note
func (n CommitNote) Total() int {
	total := 0
	for _, fm := range n.Files {
		total += fm.TimeSpent
	}
	return total
}

// Marshal converts a commit note to a serialized string
func Marshal(n CommitNote) string {
	s := fmt.Sprintf("[ver:%s,total:%d]\n", "1", n.Total())
	for _, fl := range n.Files {
		// nomralize file paths to unix convention
		s += fmt.Sprintf("%s:%d,", filepath.ToSlash(fl.SourceFile), fl.TimeSpent)
		for _, e := range fl.SortEpochs() {
			s += fmt.Sprintf("%d:%d,", e, fl.Timeline[e])
		}
		s += fmt.Sprintf("%s\n", fl.Status)
	}
	return s
}

// UnMarshal unserializes a git note string into a commit note
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
			if matches := reHeaderVals.FindAllString(lines[lineIdx], 2); len(matches) == 2 {
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
			for idx := range files {
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
	sort.Sort(sort.Reverse(FileByTime(files)))
	return CommitNote{Files: files}, nil
}

// FileDetail contains a source file's time metrics
type FileDetail struct {
	SourceFile string
	TimeSpent  int
	Timeline   map[int64]int
	Status     string
}

// ShortenSourceFile shortens source file to length n
func (f *FileDetail) ShortenSourceFile(n int) string {
	x := len(f.SourceFile) - n - 1
	if x <= 0 {
		return f.SourceFile
	}

	idx := strings.Index(f.SourceFile[x:], string(filepath.Separator))
	if idx >= 0 {
		x = x + idx
	}
	return fmt.Sprintf("...%s", f.SourceFile[x:])
}

// SortEpochs returns timeline keys sorted by epoch
func (f *FileDetail) SortEpochs() []int64 {
	keys := []int64{}
	for k := range f.Timeline {
		keys = append(keys, k)
	}
	sort.Sort(util.ByInt64(keys))
	return keys
}

// IsTerminal returns true if file is terminal
func (f *FileDetail) IsTerminal() bool {
	return f.SourceFile == ".gtm/terminal.app"
}

// IsApp returns true if file is an app event
func (f *FileDetail) IsApp() bool {
	return project.AppEventFileContentRegex.MatchString(f.SourceFile)
}

// GetAppName returns the name of the App
func (f *FileDetail) GetAppName() string {
	name := project.AppEventFileContentRegex.FindStringSubmatch(f.SourceFile)[1]
	name = util.UcFirst(name)
	return name
}

// FileByTime is list of FileDetails
type FileByTime []FileDetail

func (a FileByTime) Len() int           { return len(a) }
func (a FileByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a FileByTime) Less(i, j int) bool { return a[i].TimeSpent < a[j].TimeSpent }
