package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"

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
	tags := initFlags.String(
		"tags",
		"",
		"Set tags on project, i.e. \"work,gtm\"")
	clearTags := initFlags.Bool(
		"clear-tags",
		false,
		"Remove existing tags")
	if err := initFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}
	m, err := project.Initialize(*terminal, util.Map(strings.Split(*tags, ","), strings.TrimSpace), *clearTags)
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
