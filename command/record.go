package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/git-time-metric/gtm/event"
	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/scm"

	"github.com/mitchellh/cli"
)

type RecordCmd struct {
}

func NewRecord() (cli.Command, error) {
	return RecordCmd{}, nil
}

func (r RecordCmd) Help() string {
	return r.Synopsis()
}

func (r RecordCmd) Run(args []string) int {
	recordFlags := flag.NewFlagSet("record", flag.ExitOnError)
	status := recordFlags.Bool(
		"status",
		false,
		"After recording, return current total time spent [gtm status -total-only]")
	terminal := recordFlags.Bool(
		"terminal",
		false,
		"Record a terminal event")
	if err := recordFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	if !*terminal && len(recordFlags.Args()) == 0 {
		fmt.Println("Unable to record, file not provided")
		return 1
	}

	fileToRecord := ""
	if *terminal {
		projPath, err := scm.RootPath()
		if err != nil {
			// if not found, ignore error
			return 0
		}
		fileToRecord = filepath.Join(projPath, ".gtm", "terminal.app")
	} else {
		fileToRecord = recordFlags.Args()[0]
	}

	if err := event.Record(fileToRecord); err != nil && !(err == project.ErrNotInitialized || err == project.ErrFileNotFound) {
		if err := project.Log(err); err != nil {
			fmt.Println(err)
		}
		return 1
	} else if err == nil && *status {
		var (
			err        error
			commitNote note.CommitNote
			out        string
			wd         string
		)

		wd, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		defer os.Chdir(wd)

		os.Chdir(filepath.Dir(fileToRecord))

		if commitNote, err = metric.Process(true); err != nil {
			fmt.Println(err)
			return 1
		}
		out, err = report.Status(commitNote, report.OutputOptions{TotalOnly: true})
		if err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Printf(out)
	}

	return 0
}

func (r RecordCmd) Synopsis() string {
	return `
	Usage: gtm record [-status] [-terminal] [<path/file>]
	Record a file or terminal events

	record file event     -> gtm record /path/file
	record terminal event -> gtm record -terminal
	`
}
