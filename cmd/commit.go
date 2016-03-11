package cmd

import (
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
	metric.Process()
	return 0
}

func (r GitCommit) Synopsis() string {
	return `
	Save git commit time
	`
}
