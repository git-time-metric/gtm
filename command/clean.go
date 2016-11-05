// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/mitchellh/cli"
)

// CleanCmd contains method for clean method
type CleanCmd struct {
}

// NewClean returns a new CleanCmd struct
func NewClean() (cli.Command, error) {
	return CleanCmd{}, nil
}

// Help returns help for the clean command
func (v CleanCmd) Help() string {
	return v.Synopsis()
}

// Run executes clean command with args
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

// Synopsis return help for clean command
func (v CleanCmd) Synopsis() string {
	return `
	Usage: gtm clean [-yes]
	Cleans uncommitted time data by removing all event and metric files from the .gtm directory
	`
}
