package cmd

import (
	"flag"
	"os"

	"edgeg.io/gtm/metric"
	"github.com/mitchellh/cli"
)

type GitCommit struct {
}

func NewCommit() (cli.Command, error) {
	return GitCommit{}, nil
}

func (r GitCommit) Help() string {
	return `
	gtm commit [--dry-run] [--debug]

	Log time for git tracked files and set the file's tracked time to zero.
	`
}

func (r GitCommit) Run(args []string) int {
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	dryRun := commitFlags.Bool(
		"dry-run",
		true,
		"Do not log time but show time logged for all files")
	debug := commitFlags.Bool(
		"debug",
		false,
		"Print debug statements to the console")
	commitFlags.Parse(os.Args[2:])

	metric.Process(*dryRun, *debug)
	return 0
}

func (r GitCommit) Synopsis() string {
	return `
	Log time for git tracked files
	`
}
