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
func (c CleanCmd) Help() string {
	helpText := `
Usage: gtm clean [options]

  Deletes pending time data for the current git repository.

Options:

  -yes                       Delete time data without asking for confirmation.
`
	return strings.TrimSpace(helpText)
}

// Run executes clean command with args
func (c CleanCmd) Run(args []string) int {

	var yes bool
	cmdFlags := flag.NewFlagSet("clean", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.Usage = func() { fmt.Print(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		var response string
		fmt.Printf("\nDelete pending time data (y/n)? ")
		if _, err := fmt.Scanln(&response); err != nil {
			return 0
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		var (
			m   string
			err error
		)
		if m, err = project.Clean(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(m)
	}
	return 0
}

// Synopsis return help for clean command
func (c CleanCmd) Synopsis() string {
	return "Delete pending time data"
}
