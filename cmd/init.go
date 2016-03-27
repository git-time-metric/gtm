package cmd

import (
	"fmt"

	"edgeg.io/gtm/env"

	"github.com/mitchellh/cli"
)

type initCmd struct {
}

func NewInit() (cli.Command, error) {
	return initCmd{}, nil
}

func (i initCmd) Help() string {
	return `
	gtm init

	Initialize time tracking for a project. 
	Call from root directory of project.
	`
}

func (i initCmd) Run(args []string) int {
	err := env.Initialize()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println("Time tracking initialized")
	return 0
}

func (i initCmd) Synopsis() string {
	return `
	Initialize time tracking for a project 
	`
}
