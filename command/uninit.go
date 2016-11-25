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

// UninitCmd contains methods for uninit command
type UninitCmd struct {
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
	cmdFlags := flag.NewFlagSet("uninit", flag.ExitOnError)
	cmdFlags.BoolVar(&yes, "yes", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	confirm := yes
	if !confirm {
		var response string
		fmt.Printf("\nRemove GTM tracking for the current git repository (y/n)? ")
		_, err := fmt.Scanln(&response)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
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
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(m)
	}
	return 0
}

// Synopsis returns help for uninit command
func (c UninitCmd) Synopsis() string {
	return "Turn off time tracking for git repository"
}
