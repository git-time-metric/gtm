package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/git-time-metric/gtm/project"

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
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	terminal := initFlags.Bool(
		"terminal",
		true,
		"Track time spent in terminal (command line)")
	if err := initFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}
	m, err := project.Initialize(*terminal)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(m)
	return 0
}

func (i InitCmd) Synopsis() string {
	return `
	Usage: gtm init [-terminal=[true|false]]
	Initialize a git project for time tracking 
	`
}
