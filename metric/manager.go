package metric

import (
	"fmt"

	"github.com/git-time-metric/gtm/event"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
)

// Process events for last git commit and save time spent as a git note
// If interim is true, process events for the current working and staged files
func Process(interim bool, projPath ...string) (note.CommitNote, error) {

	rootPath, gtmPath, err := project.Paths(projPath...)
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

	commitNote := note.CommitNote{}

	if interim {
		commitMap, readonlyMap, err := buildInterimCommitMaps(metricMap, projPath...)
		if err != nil {
			fmt.Println("here")
			return note.CommitNote{}, err
		}

		commitNote, err = buildCommitNote(rootPath, metricMap, commitMap, readonlyMap)
		if err != nil {
			return note.CommitNote{}, err
		}

		// fmt.Printf("\nEventMap:\n%+v\n", epochEventMap)
		// fmt.Printf("\nMetricMap:\n%+v\n", metricMap)
		// fmt.Printf("\nCommitMap:\n%+v\n", commitMap)
		// fmt.Printf("\nReadonlyMap:\n%+v\n", readonlyMap)
	} else {
		commitMap, readonlyMap, err := buildCommitMaps(metricMap)
		if err != nil {
			return note.CommitNote{}, err
		}

		commitNote, err = buildCommitNote(rootPath, metricMap, commitMap, readonlyMap)
		if err != nil {
			return note.CommitNote{}, err
		}

		if err := scm.CreateNote(note.Marshal(commitNote), project.NoteNameSpace); err != nil {
			return note.CommitNote{}, err
		}
		if err := saveAndPurgeMetrics(gtmPath, metricMap, commitMap, readonlyMap); err != nil {
			return note.CommitNote{}, err
		}

		// fmt.Printf("\nEventMap:\n%+v\n", epochEventMap)
		// fmt.Printf("\nMetricMap:\n%+v\n", metricMap)
		// fmt.Printf("\nCommitMap:\n%+v\n", commitMap)
		// fmt.Printf("\nReadonlyMap:\n%+v\n", readonlyMap)
	}

	return commitNote, nil
}
