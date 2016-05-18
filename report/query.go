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
