package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
	Log time for git tracked files and set the file's tracked time to zero.

	gtm commit [--dry-run] [--debug]
	`
}

func (r GitCommit) Run(args []string) int {
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	yes := commitFlags.Bool(
		"yes",
		false,
		"Automatically confirm yes for commit command")
	debug := commitFlags.Bool(
		"debug",
		false,
		"Print debug statements to the console")
	commitFlags.Parse(os.Args[2:])

	confirm := *yes
	if !confirm {
		var response string
		fmt.Printf("\nSave time for last commit (y/n)? ")
		_, err := fmt.Scanln(&response)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	var (
		err error
		msg string
	)
	if confirm {
		msg, err = metric.Process(metric.Committed, *debug)
		if err != nil {
			fmt.Println(err)
			return 1
		}
	}
	fmt.Print(msg)
	return 0
}

func (r GitCommit) Synopsis() string {
	return `
	Log time for git tracked files
	`
}
