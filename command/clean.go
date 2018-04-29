// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"os"
	"strings"

	"github.com/git-time-metric/gtm/event"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

// CleanCmd contains method for clean method
type CleanCmd struct {
	Ui cli.Ui
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
  -application=true          Remove application time
  -editor=true               Remove editor time
  -terminal=true             Remove terminal time
  -days=0                    Delete starting from n days in the past
  -all=false                 Clean all projects
`
	return strings.TrimSpace(helpText)
}

// Run executes clean command with args
func (c CleanCmd) Run(args []string) int {
	var yes, terminal, application, editor, all bool
	var days int
	cmdFlags := flag.NewFlagSet("clean", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.BoolVar(&application, "application", true, "")
	cmdFlags.BoolVar(&editor, "editor", true, "")
	cmdFlags.BoolVar(&terminal, "terminal", true, "")
	cmdFlags.IntVar(&days, "days", 0, "")
	cmdFlags.BoolVar(&all, "all", false, "")
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		response, err := c.Ui.Ask("Delete pending time data (y/n)?")
		if err != nil {
			return 0
		}
		confirm = strings.TrimSpace(strings.ToLower(response)) == "y"
	}

	if confirm {
		if all {
			index, err := project.NewIndex()
			if err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			projects, err := index.Get([]string{}, all)
			if err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			for _, p := range projects {
				if err := func() error {
					d, err := os.Getwd()
					if err != nil {
						return err
					}
					defer os.Chdir(d)
					os.Chdir(p)
					err = event.Clean(util.AfterNow(days), application, editor, terminal)
					return err
				}(); err != nil {
					c.Ui.Error(err.Error())
					return 1
				}
			}
			return 0
		}

		if err := event.Clean(util.AfterNow(days), application, editor, terminal); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	}
	return 0
}

// Synopsis return help for clean command
func (c CleanCmd) Synopsis() string {
	return "Delete pending time data"
}
