// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

// CleanCmd contains method for clean method
type CleanCmd struct {
	UI cli.Ui
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
  -terminal-only             Only delete terminal time data
  -app-only                  Only delete apps time data
  -days=0                    Delete starting from n days in the past
`
	return strings.TrimSpace(helpText)
}

// Run executes clean command with args
func (c CleanCmd) Run(args []string) int {
	var yes, terminalOnly, appOnly bool
	var days int
	cmdFlags := flag.NewFlagSet("clean", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.BoolVar(&terminalOnly, "terminal-only", false, "")
	cmdFlags.BoolVar(&appOnly, "app-only", false, "")
	cmdFlags.IntVar(&days, "days", 0, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		response, err := c.UI.Ask("Delete pending time data (y/n)?")
		if err != nil {
			return 0
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		if err := project.Clean(util.AfterNow(days), terminalOnly, appOnly); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
	}
	return 0
}

// Synopsis return help for clean command
func (c CleanCmd) Synopsis() string {
	return "Delete pending time data"
}
