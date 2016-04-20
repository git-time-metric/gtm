package report

import (
	"fmt"

	"edgeg.io/gtm/commit"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

type messageLogs []messageLog

type messageLog struct {
	Message string
	Log     commit.Log
}

func retrieveLogs(commits []string) (messageLogs, error) {
	//TODO: refactor to be faster and improve error messages
	logs := messageLogs{}
	for _, c := range commits {
		n, err := scm.GitNote(c, project.NoteNameSpace)
		if err != nil {
			logs = append(logs, messageLog{Message: fmt.Sprintf("%s %s", c, err), Log: commit.Log{}})
			continue
		}
		log, err := commit.UnMarshalLog(n)
		if err != nil {
			logs = append(logs, messageLog{Message: fmt.Sprintf("%s %s", c, err), Log: commit.Log{}})
			continue
		}
		m, err := scm.GitLog(c)
		if err != nil {
			logs = append(logs, messageLog{Message: fmt.Sprintf("%s %s", c, err), Log: commit.Log{}})
			continue
		}
		logs = append(logs, messageLog{Message: m, Log: log})
	}
	return logs, nil
}
