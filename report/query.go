package report

import (
	"fmt"

	"edgeg.io/gtm/commit"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

type Logs []MessageLog

type MessageLog struct {
	Message string
	Log     commit.Log
}

func retrieveLogs(commits []string) (Logs, error) {
	logs := Logs{}
	for _, c := range commits {
		n, err := scm.GitGetNote(c, project.NoteNameSpace)
		if err != nil {
			logs = append(logs, MessageLog{Message: fmt.Sprintf("%s %s", c[:7], err), Log: commit.Log{}})
			continue
		}
		log, err := commit.UnMarshalLog(n)
		if err != nil {
			logs = append(logs, MessageLog{Message: fmt.Sprintf("%s %s", c[:7], err), Log: commit.Log{}})
			continue
		}
		m, err := scm.GitLogMessage(c)
		if err != nil {
			logs = append(logs, MessageLog{Message: fmt.Sprintf("%s %s", c[:7], err), Log: commit.Log{}})
			continue
		}
		logs = append(logs, MessageLog{Message: m, Log: log})
	}
	return logs, nil
}
