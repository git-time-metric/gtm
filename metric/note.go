package metric

import (
	"fmt"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/scm"
)

func saveNote(tl TimeLog) error {
	err := scm.GitAddNote(marshalTimeLog(tl), env.NoteNameSpace)
	if err != nil {
		return err
	}

	return nil
}

func noteForConsole(tl TimeLog) string {
	s := fmt.Sprintf("total: %d\n", tl.Total())
	for _, fl := range tl.Files {
		s += fmt.Sprintf("%s: %d [%s]\n", fl.SourceFile, fl.TimeSpent, fl.Status)
	}
	return s
}
