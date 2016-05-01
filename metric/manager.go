package metric

import (
	"fmt"

	"edgeg.io/gtm/event"
	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

func Process(gstate scm.GitState, debug bool) (note.CommitNote, error) {

	rootPath, gtmPath, err := project.Paths()
	if err != nil {
		return note.CommitNote{}, err
	}

	// load any saved metrics
	metricMap, err := loadMetrics(rootPath, gtmPath)
	if err != nil {
		return note.CommitNote{}, err
	}

	// process event files
	epochEventMap, err := event.Process(rootPath, gtmPath, gstate == scm.Working || gstate == scm.Staging)
	if err != nil {
		return note.CommitNote{}, err
	}

	// allocate time for events
	for ep := range epochEventMap {
		err := allocateTime(ep, metricMap, epochEventMap[ep])
		if err != nil {
			return note.CommitNote{}, err
		}
	}

	// build map of commit files
	commitMap, err := buildCommitMap(metricMap, gstate)
	if err != nil {
		return note.CommitNote{}, err
	}

	// create time logged struct
	logged, err := buildCommitNote(metricMap, commitMap, gstate)
	if err != nil {
		return note.CommitNote{}, err
	}

	if gstate == scm.Committed {
		if err := scm.GitAddNote(note.Marshal(logged), project.NoteNameSpace); err != nil {
			return note.CommitNote{}, err
		}
		if err := saveAndPurgeMetrics(gtmPath, metricMap, commitMap); err != nil {
			return note.CommitNote{}, err
		}
	}

	if debug {
		fmt.Printf("\nEventMap:\n%+v\n", epochEventMap)
		fmt.Printf("\nMetricMap:\n%+v\n", metricMap)
		fmt.Printf("\nCommitMap:\n%+v\n", commitMap)
	}

	return logged, nil
}
