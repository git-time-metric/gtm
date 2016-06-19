package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/report"
	"github.com/mitchellh/cli"
)

type StatusCmd struct {
}

func NewStatus() (cli.Command, error) {
	return StatusCmd{}, nil
}

func (r StatusCmd) Help() string {
	return r.Synopsis()
}

func (r StatusCmd) Run(args []string) int {
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)
	totalOnly := statusFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	if err := statusFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	var (
		err        error
		commitNote note.CommitNote
		out        string
	)

	if commitNote, err = metric.Process(true); err != nil {
		fmt.Println(err)
		return 1
	}
	out, err = report.Status(commitNote, *totalOnly)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf(out)
	return 0
}

func (r StatusCmd) Synopsis() string {
	return `
	Usage: gtm status [-total-only]
	Show time spent for working or staged files
	`
}
