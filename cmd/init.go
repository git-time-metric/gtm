package cmd

import (
	"fmt"

	"edgeg.io/gtm/cfg"

	"github.com/mitchellh/cli"
)

type initCmd struct {
}

func NewInit() (cli.Command, error) {
	return initCmd{}, nil
}

func (i initCmd) Help() string {
	return `
	Initialize Git Metric to start recording time

	The init command is required to be called from
	the root of the git project.
	`
}

func (i initCmd) Run(args []string) int {
	err := cfg.Initialize()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func (i initCmd) Synopsis() string {
	return `
	Initialize Git Metric to start recording time
	`
}
