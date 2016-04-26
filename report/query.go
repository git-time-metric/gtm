package report

import (
	"fmt"

	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

type commitNoteDetails []commitNoteDetail

type commitNoteDetail struct {
	Author  string
	Date    string
	Hash    string
	Subject string
	Note    note.CommitNote
}

func retrieveNotes(commits []string) (commitNoteDetails, error) {
	//TODO: refactor to be faster and improve error messages
	notes := commitNoteDetails{}
	for _, c := range commits {
		n, err := scm.GitNote(c, project.NoteNameSpace)
		if err != nil {
			notes = append(notes, commitNoteDetail{Subject: fmt.Sprintf("%s %s", c, err), Note: note.CommitNote{}})
			continue
		}
		log, err := note.UnMarshal(n)
		if err != nil {
			notes = append(notes, commitNoteDetail{Subject: fmt.Sprintf("%s %s", c, err), Note: note.CommitNote{}})
			continue
		}
		fields, err := scm.GitLog(c)
		if err != nil {
			notes = append(notes, commitNoteDetail{Subject: fmt.Sprintf("%s %s", c, err), Note: note.CommitNote{}})
			continue
		}
		notes = append(notes, commitNoteDetail{Author: fields[0], Date: fields[1], Hash: fields[2], Subject: fields[3], Note: log})
	}
	return notes, nil
}
