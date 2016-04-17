package report

import (
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
			return Logs{}, err
		}
		log, err := commit.UnMarshalLog(n)
		if err != nil {
			return Logs{}, err
		}
		m, err := scm.GitLogMessage(c)
		if err != nil {
			return Logs{}, err
		}
		logs = append(logs, MessageLog{Message: m, Log: log})
	}
	return logs, nil
}
