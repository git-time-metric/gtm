// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/metric"
	"github.com/mitchellh/cli"
)

// GitCommit struct contain methods for commit command
type GitCommit struct {
}

// NewCommit returns new GitCommit struct
func NewCommit() (cli.Command, error) {
	return GitCommit{}, nil
}

// Help returns help for commit command
func (c GitCommit) Help() string {
	helpText := `
Usage: gtm commit [options]

  Save pending time with last commit. 

Options:

  -yes                       Save time data without asking for confirmation.
`
	return strings.TrimSpace(helpText)
}

// Run executes commit commands with args
func (c GitCommit) Run(args []string) int {

	var yes bool
	cmdFlags := flag.NewFlagSet("commit", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.Usage = func() { fmt.Print(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		var response string
		fmt.Printf("\nSave time for last commit (y/n)? ")
		if _, err := fmt.Scanln(&response); err != nil {
			return 0
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		if _, err := metric.Process(false); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

// Synopsis return help for commit command
func (c GitCommit) Synopsis() string {
	return "Save pending time with the last commit"
}
