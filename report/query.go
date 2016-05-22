package report

import (
	"sort"
	"strings"
	"time"

	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
	"edgeg.io/gtm/util"
)

func retrieveNotes(commits []string) commitNoteDetails {
	notes := commitNoteDetails{}
	for _, c := range commits {
		if len(c) > 7 {
			c = c[:7]
		}
		gitFlds, err := scm.GitLog(c)
		if err != nil {
			notes = append(notes, commitNoteDetail{Hash: c, Subject: "Invalid sha1\n", Note: note.CommitNote{}})
			continue
		}
		noteText, err := scm.GitNote(c, project.NoteNameSpace)
		if err != nil {
			notes = append(
				notes,
				commitNoteDetail{Author: gitFlds[0], Date: gitFlds[1], Hash: gitFlds[2], Subject: gitFlds[3], Note: note.CommitNote{}})
			continue
		}
		commitNote, err := note.UnMarshal(noteText)
		if err != nil {
			notes = append(
				notes,
				commitNoteDetail{Author: gitFlds[0], Date: gitFlds[1], Hash: gitFlds[2], Subject: gitFlds[3], Note: note.CommitNote{}})
			continue
		}
		notes = append(notes, commitNoteDetail{Author: gitFlds[0], Date: gitFlds[1], Hash: gitFlds[2], Subject: gitFlds[3], Note: commitNote})
	}
	return notes
}

type commitNoteDetails []commitNoteDetail

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
	Hash    string
	Subject string
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
	total := 0
	for _, entry := range f {
		total += entry.Seconds
	}
	return util.FormatDuration(total)
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
