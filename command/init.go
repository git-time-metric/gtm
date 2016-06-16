package command

import (
	"fmt"

	"edgeg.io/gtm/project"

	"github.com/mitchellh/cli"
)

type InitCmd struct {
}

func NewInit() (cli.Command, error) {
	return InitCmd{}, nil
}

func (i InitCmd) Help() string {
	return i.Synopsis()
}

func (i InitCmd) Run(args []string) int {
	m, err := project.Initialize()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(m)
	return 0
}

func (i InitCmd) Synopsis() string {
	return `
	Usage: gtm init
	Initialize a git project for time tracking 
	`
}
