// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

const (
	defaultDateFormat = "Mon Jan 02 15:04:05 2006 MST"
)

func retrieveNotes(projects []ProjectCommits, terminalOff, appOff, calcStats bool, dateFormat string) commitNoteDetails {
	notes := commitNoteDetails{}

	if dateFormat == "" {
		dateFormat = defaultDateFormat
	}

	for _, p := range projects {
		for _, c := range p.Commits {

			n, err := scm.ReadNote(c, project.NoteNameSpace, calcStats, p.Path)
			if err != nil {
				notes = append(notes, commitNoteDetail{})
				continue
			}

			when := n.When.Format(dateFormat)

			var commitNote note.CommitNote
			commitNote, err = note.UnMarshal(n.Note)
			if err != nil {
				commitNote = note.CommitNote{}
			}

			if terminalOff {
				commitNote = commitNote.FilterOutTerminal()
			}
			if appOff {
				commitNote = commitNote.FilterOutApp()
			}

			id := n.ID
			if len(id) > 7 {
				id = id[:7]
			}

			message := strings.TrimPrefix(n.Message, n.Summary)
			message = strings.TrimSpace(message)

			notes = append(notes,
				commitNoteDetail{
					Author:     n.Author,
					Date:       when,
					When:       n.When,
					Hash:       id,
					Subject:    n.Summary,
					Message:    message,
					Note:       commitNote,
					Project:    filepath.Base(p.Path),
					LineAdd:    fmt.Sprintf("+%d", n.Stats.Insertions),
					LineDel:    fmt.Sprintf("-%d", n.Stats.Deletions),
					LineDiff:   fmt.Sprintf("%d", n.Stats.Insertions-n.Stats.Deletions),
					ChangeRate: fmt.Sprintf("%.0f", n.Stats.ChangeRatePerHour(commitNote.Total())),
				})
		}
	}
	sort.Sort(notes)
	return notes
}

type commitNoteDetails []commitNoteDetail

func (c commitNoteDetails) Len() int           { return len(c) }
func (c commitNoteDetails) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c commitNoteDetails) Less(i, j int) bool { return c[i].When.After(c[j].When) }

func (c commitNoteDetails) Total() int {
	t := 0
	for i := range c {
		t += c[i].Note.Total()
	}
	return t
}

type commitNoteDetail struct {
	Author     string
	Date       string
	When       time.Time
	Hash       string
	Subject    string
	Project    string
	Message    string
	Note       note.CommitNote
	LineAdd    string
	LineDel    string
	LineDiff   string
	ChangeRate string
}

func (c commitNoteDetails) files() fileEntries {
	filesMap := map[string]fileEntry{}
	for _, n := range c {
		for _, f := range n.Note.Files {
			if entry, ok := filesMap[f.SourceFile]; !ok {
				filesMap[f.SourceFile] = fileEntry{Filename: f.SourceFile, Seconds: f.TimeSpent}
			} else {
				entry.add(f.TimeSpent)
				filesMap[f.SourceFile] = entry
			}
		}
	}

	files := make(fileEntries, 0, len(filesMap))
	for _, entry := range filesMap {
		files = append(files, entry)
	}
	sort.Sort(sort.Reverse(files))
	return files
}

type fileEntries []fileEntry

func (f fileEntries) Len() int           { return len(f) }
func (f fileEntries) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f fileEntries) Less(i, j int) bool { return f[i].Seconds < f[j].Seconds }

func (f fileEntries) Duration() string {
	return util.FormatDuration(f.Total())
}

func (f fileEntries) Total() int {
	total := 0
	for _, entry := range f {
		total += entry.Seconds
	}
	return total
}

type fileEntry struct {
	Filename string
	Seconds  int
}

func (f *fileEntry) add(s int) {
	f.Seconds += s
}

func (f *fileEntry) Duration() string {
	return util.FormatDuration(f.Seconds)
}

func (f *fileEntry) IsTerminal() bool {
	return f.Filename == ".gtm/terminal.app"
}

func (f *fileEntry) IsApp() bool {
	return project.AppEventFileContentRegex.MatchString(f.Filename)
}

// GetAppName returns the name of the App
func (f *fileEntry) GetAppName() string {
	name := project.AppEventFileContentRegex.FindStringSubmatch(f.Filename)[1]
	name = util.UcFirst(name)
	return name
}
