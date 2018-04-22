// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package command

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/event"
	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/util"

	"github.com/mitchellh/cli"
)

// RecordCmd contains method for record command
type RecordCmd struct {
	Ui  cli.Ui
	Out *bytes.Buffer
}

func (c RecordCmd) output(s string) {
	if c.Out != nil {
		fmt.Fprint(c.Out, s)
	} else {
		fmt.Fprint(os.Stdout, s)
	}
}

// NewRecord return new RecordCmd struct
func NewRecord() (cli.Command, error) {
	return RecordCmd{}, nil
}

// Help returns help for record command
func (c RecordCmd) Help() string {
	helpText := `
Usage: gtm record [options] [/path/file]

  Record file or terminal events.

Options:

  -terminal=false            Record a terminal event.

  -app=""                    Record an application event.

  -status=false              Return total time recorded for event.

  -long-duration=false       Return total time recorded in long duration format
`
	return strings.TrimSpace(helpText)
}

// Run executes record command with args
func (c RecordCmd) Run(args []string) int {
	defer util.TimeTrack(time.Now(), "command.Record")

	var status, terminal, longDuration bool
	var application string
	cmdFlags := flag.NewFlagSet("record", flag.ContinueOnError)
	cmdFlags.BoolVar(&status, "status", false, "")
	cmdFlags.BoolVar(&terminal, "terminal", false, "")
	cmdFlags.BoolVar(&longDuration, "long-duration", false, "")
	cmdFlags.StringVar(&application, "app", "", "")
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !terminal && application == "" && len(cmdFlags.Args()) == 0 {
		c.Ui.Error("Unable to record, file not provided")
		return 1
	}

	// TODO: test performance of turning of status for record
	// Checking status on record is expensive, +40ms with and +2ms without
	// Default terminal plugin to not check status on record

	// status = false

	outputStatus := func(path string) int {
		if status {
			var (
				err        error
				commitNote note.CommitNote
				out        string
				wd         string
			)

			wd, err = os.Getwd()
			if err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			defer os.Chdir(wd)

			os.Chdir(filepath.Dir(path))

			if commitNote, err = metric.Process(true); err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			out, err = report.Status(commitNote, report.OutputOptions{TotalOnly: true, LongDuration: longDuration})
			if err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			c.output(out)
		}
		return 0
	}

	switch {
	case terminal:
		// terminal plugin
		a, err := event.NewTerminalApplication()
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		// we want terminal events to update the active project
		// we do this by using event.Record() instead a.Record()
		if err := event.Record(a.Path()); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		return outputStatus(a.Path())

	case application != "":
		a, err := event.NewApplicationFromName(application)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		if err := a.Record(); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		return outputStatus(a.Path())

	default:
		err := event.Record(cmdFlags.Args()[0])
		if err != nil && !(err == project.ErrNotInitialized || err == project.ErrFileNotFound) {
			return 1
		}
		return outputStatus(cmdFlags.Args()[0])
	}

	return 0
}

// Synopsis returns help
func (c RecordCmd) Synopsis() string {
	return "Record file and terminal events"
}
