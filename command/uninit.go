// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"strings"

	"github.com/kilpkonn/gtm/project"
	"github.com/mitchellh/cli"
)

// UninitCmd contains methods for uninit command
type UninitCmd struct {
	UI cli.Ui
}

// NewUninit returns new UninitCmd struct
func NewUninit() (cli.Command, error) {
	return UninitCmd{}, nil
}

// Help returns help for uninit command
func (c UninitCmd) Help() string {
	helpText := `
Usage: gtm uninit [options]

  Turn off time tracking for git repository (does not remove committed time data).

Options:

  -yes                       Turn off without asking for confirmation.
`
	return strings.TrimSpace(helpText)
}

// Run executes uninit command with args
func (c UninitCmd) Run(args []string) int {
	var yes bool
	cmdFlags := flag.NewFlagSet("uninit", flag.ContinueOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		var response string
		response, err := c.UI.Ask("Remove GTM tracking for the current git repository (y/n)?")
		if err != nil {
			c.UI.Error(err.Error())
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
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Output(m)
	}
	return 0
}

// Synopsis returns help for uninit command
func (c UninitCmd) Synopsis() string {
	return "Turn off time tracking for git repository"
}
