package metric

import (
	"fmt"

	"edgeg.io/gtm/event"
	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/scm"
)

// Process events for last git commit and save time spent as a git note
// If interim is true, process events for the current working and staged files
func Process(interim bool, debug bool) (note.CommitNote, error) {

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
	epochEventMap, err := event.Process(rootPath, gtmPath, interim)
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
	commitMap, err := buildCommitMap(metricMap)
	if err != nil {
		return note.CommitNote{}, err
	}

	// build commit note struct
	commitNote, err := buildCommitNote(metricMap, commitMap)
	if err != nil {
		return note.CommitNote{}, err
	}

	if !interim {
		if err := scm.GitAddNote(note.Marshal(commitNote), project.NoteNameSpace); err != nil {
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

	return commitNote, nil
}
