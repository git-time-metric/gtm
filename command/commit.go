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
	return r.Synopsis()
}

func (r GitCommit) Run(args []string) int {
	commitFlags := flag.NewFlagSet("commit", flag.ExitOnError)
	yes := commitFlags.Bool(
		"yes",
		false,
		"Automatically confirm yes for saving logged time with last commit")
	if err := commitFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

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

	if confirm {
		if _, err := metric.Process(false); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	return 0
}

func (r GitCommit) Synopsis() string {
	return `
	Usage: gtm commit [--yes]
	Save your logged time with the last commit
	This is automatically called from the postcommit hook
	Warning - any time logged will be cleared from your working directory
	`
}
