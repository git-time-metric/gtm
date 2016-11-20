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
	"github.com/git-time-metric/gtm/util"

	"github.com/mitchellh/cli"
)

// InitCmd contains methods for init command
type InitCmd struct {
}

// NewInit returns new InitCmd struct
func NewInit() (cli.Command, error) {
	return InitCmd{}, nil
}

// Help return help for init command
func (i InitCmd) Help() string {
	return i.Synopsis()
}

// Run executes init command with args
func (i InitCmd) Run(args []string) int {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	terminal := initFlags.Bool(
		"terminal",
		true,
		"Track time spent in terminal (command line)")
	tags := initFlags.String(
		"tags",
		"",
		"Set tags on project, i.e. \"work,gtm\"")
	clearTags := initFlags.Bool(
		"clear-tags",
		false,
		"Remove existing tags")
	if err := initFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	m, err := project.Initialize(*terminal, util.Map(strings.Split(*tags, ","), strings.TrimSpace), *clearTags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(m)
	return 0
}

// Synopsis return help for init command
func (i InitCmd) Synopsis() string {
	return `
	Usage: gtm init [-terminal=[true|false]] [-tags tag1,tag2] [-clear-tags]
	Initialize a git project for time tracking 
	`
}
