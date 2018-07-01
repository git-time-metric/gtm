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

// InitCmd contains methods for init command
type InitCmd struct {
	UI cli.Ui
}

// NewInit returns new InitCmd struct
func NewInit() (cli.Command, error) {
	return InitCmd{}, nil
}

// Help return help for init command
func (c InitCmd) Help() string {
	helpText := `
Usage: gtm init [options]

  Initialize a git repository for time tracking.

Options:

  -terminal=true             Enable time tracking for terminal (requires Terminal plug-in).

  -tags=tag1,tag2            Add tags to projects, multiple calls appends tags.

  -clear-tags                Clear all tags.
`
	return strings.TrimSpace(helpText)
}

// Run executes init command with args
func (c InitCmd) Run(args []string) int {
	var terminal, clearTags bool
	var tags string
	cmdFlags := flag.NewFlagSet("init", flag.ContinueOnError)
	cmdFlags.BoolVar(&terminal, "terminal", true, "")
	cmdFlags.BoolVar(&clearTags, "clear-tags", false, "")
	cmdFlags.StringVar(&tags, "tags", "", "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	m, err := project.Initialize(terminal, util.Map(strings.Split(tags, ","), strings.TrimSpace), clearTags)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	c.UI.Output(m + "\n")
	return 0
}

// Synopsis return help for init command
func (c InitCmd) Synopsis() string {
	return "Initialize git repository for time tracking"
}
