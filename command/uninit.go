package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/mitchellh/cli"
)

// UninitCmd contains methods for uninit command
type UninitCmd struct {
}

// NewUninit returns new UninitCmd struct
func NewUninit() (cli.Command, error) {
	return UninitCmd{}, nil
}

// Help returns help for uninit command
func (v UninitCmd) Help() string {
	return v.Synopsis()
}

// Run executes uninit command with args
func (v UninitCmd) Run(args []string) int {

	uninitFlags := flag.NewFlagSet("uninit", flag.ExitOnError)
	yes := uninitFlags.Bool(
		"yes",
		false,
		"Automatically confirm yes to remove GTM tracking for the current Git repository")
	if err := uninitFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	confirm := *yes
	if !confirm {
		var response string
		fmt.Printf("\nRemove GTM tracking for the current git repository (y/n)? ")
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
		if m, err = project.Uninitialize(); err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Println(m)
	}
	return 0
}

// Synopsis returns help for uninit command
func (v UninitCmd) Synopsis() string {
	return `
	Usage: gtm uninit [-yes]
	Remove GTM tracking for the current git repository 
	Note - this removes uncommitted time data but does not remove time data that is committed
	`
}
