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

	epochEventMap, err := event.Process(gtmPath, dryRun)
	if err != nil {
		return err
	}

	metricMap, err := loadMetrics(gtmPath)
	if err != nil {
		return err
	}

	for epoch := range epochEventMap {
		err := allocateTime(metricMap, epochEventMap[epoch])
		if err != nil {
			return err
		}
	}

	m, err := scm.GitCommitMsg()
	if err != nil {
		return err
	}
	_, _, commitFiles := scm.GitParseMessage(m)

	commitMap := map[string]metricFile{}
	if !dryRun {
		//for only files in the last commit
		for _, f := range commitFiles {
			fileID := getFileID(f)
			if _, ok := metricMap[fileID]; !ok {
				continue
			}
			commitMap[fileID] = metricMap[fileID]
		}
	}

	if err := writeNote(gtmPath, metricMap, commitMap, dryRun); err != nil {
		return err
	}
	if err := saveMetrics(gtmPath, metricMap, commitMap, dryRun); err != nil {
		return err
	}

	if debug {
		fmt.Printf("\nEventMap:\n%+v\n", epochEventMap)
		fmt.Printf("\nMetricMap:\n%+v\n", metricMap)
		fmt.Printf("\nCommitMap:\n%+v\n", commitMap)
	}

	return nil
}
