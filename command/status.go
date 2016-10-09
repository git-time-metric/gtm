package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
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
	color := statusFlags.Bool(
		"color",
		false,
		"Always output color even if no terminal is detected. Use this with pagers i.e 'less -R' or 'more -R'")
	terminalOff := statusFlags.Bool(
		"terminal-off",
		false,
		"Exclude time spent in terminal (Terminal plugin is required)")
	totalOnly := statusFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	tags := statusFlags.String(
		"tags",
		"",
		"Project tags to show status on")
	all := statusFlags.Bool(
		"all",
		false,
		"Show status for all projects")
	if err := statusFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	if *totalOnly && (*all || *tags != "") {
		fmt.Fprint(os.Stderr, "\n-tags and -all options not allowed with -total-only\n")
		return 1
	}

	var (
		err        error
		commitNote note.CommitNote
		out        string
	)

	index, err := project.NewIndex()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	projects, err := index.Get(strings.Fields(*tags), *all)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, projPath := range projects {
		if commitNote, err = metric.Process(true, projPath); err != nil {
			fmt.Fprint(os.Stderr, err)
			return 1
		}
		o, err := report.Status(commitNote, *totalOnly, *terminalOff, *color, projPath)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return 1
		}
		out += o
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
