package command

import (
	"fmt"

	"edgeg.io/gtm/project"

	"github.com/mitchellh/cli"
)

type initCmd struct {
}

func NewInit() (cli.Command, error) {
	return initCmd{}, nil
}

func (i initCmd) Help() string {
	return `
	Initialize Git Time Metric for a project 

	gtm init

	Note - run from the root directory of project
	`
}

func (i initCmd) Run(args []string) int {
	m, err := project.Initialize()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(m)
	return 0
}

func (i initCmd) Synopsis() string {
	return `
	Initialize Git Time Metric for a project 
	`
}
