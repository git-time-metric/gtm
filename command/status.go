package command

import (
	"flag"
	"fmt"
	"os"

	"edgeg.io/gtm/metric"
	"edgeg.io/gtm/report"
	"github.com/mitchellh/cli"
)

type StatusCmd struct {
}

func NewStatus() (cli.Command, error) {
	return StatusCmd{}, nil
}

func (r StatusCmd) Help() string {
	return `
	Show time spent for working or staged files
	`
}

func (r StatusCmd) Run(args []string) int {
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)
	totalOnly := statusFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	statusFlags.Parse(os.Args[2:])
	out := ""
	if commitNote, err := metric.Process(true); err != nil {
		fmt.Println(err)
		return 1
	} else {
		if *totalOnly {
			out = report.NoteFilesTotal(commitNote)
		} else {
			out, err = report.NoteFiles(commitNote)
			if err != nil {
				fmt.Println(err)
				return 1
			}
		}
	}
	fmt.Printf(out)
	return 0
}

func (r StatusCmd) Synopsis() string {
	return `
	Show time spent for working or staged files
	`
}
