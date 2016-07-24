package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/mitchellh/cli"
)

type CleanCmd struct {
}

func NewClean() (cli.Command, error) {
	return CleanCmd{}, nil
}

func (v CleanCmd) Help() string {
	return v.Synopsis()
}

func (v CleanCmd) Run(args []string) int {

	cleanFlags := flag.NewFlagSet("clean", flag.ExitOnError)
	yes := cleanFlags.Bool(
		"yes",
		false,
		"Automatically confirm yes for cleaning uncommitted time data")
	if err := cleanFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	confirm := *yes
	if !confirm {
		var response string
		fmt.Printf("\nClean uncommitted time data (y/n)? ")
		_, err := fmt.Scanln(&response)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		var (
			m   string
			err error
		)
		if m, err = project.Clean(); err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Println(m)
	}
	return 0
}

func (v CleanCmd) Synopsis() string {
	return `
	Usage: gtm clean [-yes]
	Cleans uncommitted time data by removing all event and metric files from the .gtm directory
	`
}
