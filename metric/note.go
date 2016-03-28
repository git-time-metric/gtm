package metric

import (
	"fmt"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/scm"
)

func saveNote(tl timeLogged) error {
	// TODO: implement marshal and unmarshal note data
	err := scm.GitAddNote(noteForConsole(tl), env.NoteNameSpace)
	if err != nil {
		return err
	}

	return nil
}

func noteForConsole(tl timeLogged) string {
	s := fmt.Sprintf("Total: %d\n", tl.Total())
	for _, mf := range tl.Files {
		s += fmt.Sprintf("%s: %d [%s]\n", mf.SourceFile, mf.TimeSpent, tl.FileStatus(mf.SourceFile))
	}
	return s
}
