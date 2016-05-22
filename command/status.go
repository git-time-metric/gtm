package command

import (
	"flag"
	"fmt"
	"os"

	"edgeg.io/gtm/metric"
	"edgeg.io/gtm/note"
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
	Show time spent for working or staged files
	`
}
