// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/git-time-metric/gtm/command"
	"github.com/mitchellh/cli"
)

// Version is the released version set during the release build process
var Version = "0.0.0"

func main() {
	ui := &cli.ColoredUi{ErrorColor: cli.UiColorRed, Ui: &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin}}
	c := cli.NewCLI("gtm", Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init": func() (cli.Command, error) {
			return &command.InitCmd{
				Ui: ui,
			}, nil
		},
		"record": func() (cli.Command, error) {
			return &command.RecordCmd{
				Ui: ui,
			}, nil
		},
		"commit": func() (cli.Command, error) {
			return &command.CommitCmd{
				Ui: ui,
			}, nil
		},
		"report": func() (cli.Command, error) {
			return &command.ReportCmd{
				Ui: ui,
			}, nil
		},
		"status": func() (cli.Command, error) {
			return &command.StatusCmd{
				Ui: ui,
			}, nil
		},
		"verify": func() (cli.Command, error) {
			return &command.VerifyCmd{
				Ui:      ui,
				Version: Version,
			}, nil
		},
		"uninit": func() (cli.Command, error) {
			return &command.UninitCmd{
				Ui: ui,
			}, nil
		},
		"clean": func() (cli.Command, error) {
			return &command.CleanCmd{
				Ui: ui,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
