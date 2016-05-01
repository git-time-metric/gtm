package report

import (
	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

type commitNoteDetails []commitNoteDetail

func (notes commitNoteDetails) Total() int {
	t := 0
	for i := range notes {
		t += notes[i].Note.Total()
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

func retrieveNotes(commits []string) commitNoteDetails {
	notes := commitNoteDetails{}
	for _, c := range commits {
		n, err := scm.GitNote(c, project.NoteNameSpace)
		msg := "No Git Time Metric data"
		if len(c) > 7 {
			c = c[:7]
		}
		if err != nil {
			notes = append(notes, commitNoteDetail{Hash: c, Subject: msg, Note: note.CommitNote{}})
			continue
		}
		log, err := note.UnMarshal(n)
		if err != nil {
			notes = append(notes, commitNoteDetail{Hash: c, Subject: msg, Note: note.CommitNote{}})
			continue
		}
		fields, err := scm.GitLog(c)
		if err != nil {
			notes = append(notes, commitNoteDetail{Hash: c, Subject: msg, Note: note.CommitNote{}})
			continue
		}
		notes = append(notes, commitNoteDetail{Author: fields[0], Date: fields[1], Hash: fields[2], Subject: fields[3], Note: log})
	}
	return notes
}
