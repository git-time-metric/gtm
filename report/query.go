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

type commitNoteDetails []commitNoteDetail

func (notes commitNoteDetails) Total() int {
	t := 0
	for i := range notes {
		t += notes[i].Note.Total()
	}
	return t
}

type timeline []TimelineEntry

func (t timeline) Duration() string {
	total := 0
	for _, entry := range t {
		total += entry.Seconds
	}
	return util.FormatDuration(total)
}

type TimelineEntry struct {
	Day     string
	Seconds int
}

func (t *TimelineEntry) Add(s int) {
	t.Seconds += s
}

func (t *TimelineEntry) Bars() string {
	if t.Seconds == 0 {
		return ""
	}
	return strings.Repeat("*", 1+(t.Seconds/3601))
}

func (t *TimelineEntry) Duration() string {
	return util.FormatDuration(t.Seconds)
}

func (notes commitNoteDetails) Timeline() timeline {
	timelineMap := map[string]TimelineEntry{}
	timeline := []TimelineEntry{}
	for _, n := range notes {
		for _, f := range n.Note.Files {
			for epoch, secs := range f.Timeline {
				t := time.Unix(epoch, 0)
				day := t.Format("2006-01-02")
				if entry, ok := timelineMap[day]; !ok {
					timelineMap[day] = TimelineEntry{Day: t.Format("Mon Jan 02"), Seconds: secs}
				} else {
					entry.Add(secs)
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

type commitNoteDetail struct {
	Author  string
	Date    string
	Hash    string
	Subject string
	Note    note.CommitNote
}

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
