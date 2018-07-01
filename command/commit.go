// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"strings"

	"github.com/git-time-metric/gtm/metric"
	"github.com/mitchellh/cli"
)

// CommitCmd struct contain methods for commit command
type CommitCmd struct {
	UI cli.Ui
}

// NewCommit returns new CommitCmd struct
func NewCommit() (cli.Command, error) {
	return CommitCmd{}, nil
}

// Help returns help for commit command
func (c CommitCmd) Help() string {
	helpText := `
Usage: gtm commit [options]

  Save pending time with last commit. 

Options:

  -yes                       Save time data without asking for confirmation.
`
	return strings.TrimSpace(helpText)
}

// Run executes commit commands with args
func (c CommitCmd) Run(args []string) int {

	var yes bool
	cmdFlags := flag.NewFlagSet("commit", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		response, err := c.UI.Ask("Save time for last commit (y/n)?")
		if err != nil {
			return 0
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		if _, err := metric.Process(false); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
	}
	return 0
}

// Synopsis return help for commit command
func (c CommitCmd) Synopsis() string {
	return "Save pending time with the last commit"
}
