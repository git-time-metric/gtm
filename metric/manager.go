package metric

import (
	"fmt"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/event"
	"edgeg.io/gtm/scm"
)

func Process(dryRun, debug bool) error {
	_, gtmPath, err := env.Paths()
	if err != nil {
		return err
	}

	// load any saved metrics
	metricMap, err := loadMetrics(gtmPath)
	if err != nil {
		return err
	}

	// process event files
	epochEventMap, err := event.Process(gtmPath, dryRun)
	if err != nil {
		return err
	}

	// allocate time for events
	for epoch := range epochEventMap {
		err := allocateTime(metricMap, epochEventMap[epoch])
		if err != nil {
			return err
		}
	}

	// build map of commit files
	commitMap, err := buildCommitMap(metricMap, dryRun)
	if err != nil {
		return err
	}

	// create time logged struct
	logged, err := newTimeLogged(metricMap, commitMap)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println(noteForConsole(logged))
	} else {
		if err := saveNote(logged); err != nil {
			return err
		}
		if err := saveMetrics(gtmPath, metricMap, commitMap); err != nil {
			return err
		}
	}

	if debug {
		fmt.Printf("\nEventMap:\n%+v\n", epochEventMap)
		fmt.Printf("\nMetricMap:\n%+v\n", metricMap)
		fmt.Printf("\nCommitMap:\n%+v\n", commitMap)
	}

	return nil
}

func buildCommitMap(metricMap map[string]metricFile, dryRun bool) (map[string]metricFile, error) {
	commitMap := map[string]metricFile{}

	if !dryRun {
		// for only files in the last commit
		m, err := scm.GitCommitMsg()
		if err != nil {
			return commitMap, err
		}
		_, _, commitFiles := scm.GitParseMessage(m)
		for _, f := range commitFiles {
			fileID := getFileID(f)
			if _, ok := metricMap[fileID]; !ok {
				continue
			}
			commitMap[fileID] = metricMap[fileID]
		}
	} else {
		// include git tracked files that have been modified
		for fileID, mf := range metricMap {
			if mf.GitTracked {
				modified, err := scm.GitModified(mf.SourceFile)
				if err != nil {
					return commitMap, err
				}
				if modified {
					commitMap[fileID] = mf
				}
			}
		}
	}

	return commitMap, nil
}
