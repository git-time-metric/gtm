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
	Save git commit time

	gmetric commit
	`
}

func (r GitCommit) Run(args []string) int {
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	dryRun := commitFlags.Bool(
		"dry-run",
		true,
		"Do not create a note for the last commit and clear time metrics")
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
	Save git commit time
	`
}
