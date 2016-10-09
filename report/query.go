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

func retrieveNotes(projects []ProjectCommits, terminal bool) commitNoteDetails {
	notes := commitNoteDetails{}

	for _, p := range projects {
		for _, c := range p.Commits {

			n, err := scm.ReadNote(c, project.NoteNameSpace, p.Path)
			if err != nil {
				notes = append(notes, commitNoteDetail{})
				continue
			}

			when := n.When.Format("Mon Jan 02 15:04:05 2006 MST")

			var commitNote note.CommitNote
			commitNote, err = note.UnMarshal(n.Note)
			if err != nil {
				project.Log(fmt.Sprintf("Error unmarshalling note \n\n%s \n\n%s", n.Note, err))
				commitNote = note.CommitNote{}
			}

			if !terminal {
				commitNote = commitNote.FilterOutTerminal()
			}

			id := n.ID
			if len(id) > 7 {
				id = id[:7]
			}

			message := strings.TrimPrefix(n.Message, n.Summary)
			message = strings.TrimSpace(message)

			notes = append(notes,
				commitNoteDetail{
					Author:  n.Author,
					Date:    when,
					When:    n.When,
					Hash:    id,
					Subject: n.Summary,
					Message: message,
					Note:    commitNote,
					Project: filepath.Base(p.Path),
				})
		}
	}
	sort.Sort(notes)
	return notes
}

type commitNoteDetails []commitNoteDetail

func (n commitNoteDetails) Len() int           { return len(n) }
func (n commitNoteDetails) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n commitNoteDetails) Less(i, j int) bool { return n[i].When.After(n[j].When) }

func (n commitNoteDetails) Total() int {
	t := 0
	for i := range n {
		t += n[i].Note.Total()
	}
	return t
}

type commitNoteDetail struct {
	Author  string
	Date    string
	When    time.Time
	Hash    string
	Subject string
	Project string
	Message string
	Note    note.CommitNote
}

func (n commitNoteDetails) timeline() timelineEntries {
	timelineMap := map[string]timelineEntry{}
	timeline := []timelineEntry{}
	for _, n := range n {
		for _, f := range n.Note.Files {
			for epoch, secs := range f.Timeline {
				t := time.Unix(epoch, 0)
				day := t.Format("2006-01-02")
				if entry, ok := timelineMap[day]; !ok {
					timelineMap[day] = timelineEntry{Day: t.Format("Mon Jan 02"), Seconds: secs}
				} else {
					entry.add(secs)
					timelineMap[day] = entry
				}
			}
		}
	}

	keys := make([]string, 0, len(timelineMap))
	for key := range timelineMap {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, k := range keys {
		timeline = append(timeline, timelineMap[k])
	}
	return timeline
}

type timelineEntries []timelineEntry

func (t timelineEntries) Duration() string {
	total := 0
	for _, entry := range t {
		total += entry.Seconds
	}
	return util.FormatDuration(total)
}

type timelineEntry struct {
	Day     string
	Seconds int
}

func (t *timelineEntry) add(s int) {
	t.Seconds += s
}

func (t *timelineEntry) Bars() string {
	if t.Seconds == 0 {
		return ""
	}
	return strings.Repeat("*", 1+(t.Seconds/3601))
}

func (t *timelineEntry) Duration() string {
	return util.FormatDuration(t.Seconds)
}

func (n commitNoteDetails) files() fileEntries {
	filesMap := map[string]fileEntry{}
	for _, n := range n {
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
