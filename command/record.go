// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-time-metric/gtm/event"
	"github.com/git-time-metric/gtm/metric"
	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/scm"

	"github.com/mitchellh/cli"
)

// RecordCmd contains method for record command
type RecordCmd struct {
	UI  cli.Ui
	Out *bytes.Buffer
}

func (c RecordCmd) output(s string) {
	var err error
	if c.Out != nil {
		_, err = fmt.Fprint(c.Out, s)
	} else {
		_, err = fmt.Fprint(os.Stdout, s)
	}
	if err != nil {
		fmt.Printf("Error printing output, %s\n", err)
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

  Record file or app events.

Options:

  -terminal=false            Record a terminal event.

  -status=false              Return total time recorded for event.

  -long-duration=false       Return total time recorded in long duration format.

  -app=false                 Record an app event.
`
	return strings.TrimSpace(helpText)
}

// Run executes record command with args
func (c RecordCmd) Run(args []string) int {
	var status, terminal, longDuration, app bool
	cmdFlags := flag.NewFlagSet("record", flag.ContinueOnError)
	cmdFlags.BoolVar(&status, "status", false, "")
	cmdFlags.BoolVar(&terminal, "terminal", false, "")
	cmdFlags.BoolVar(&longDuration, "long-duration", false, "")
	cmdFlags.BoolVar(&app, "app", false, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !terminal && len(cmdFlags.Args()) == 0 {
		c.UI.Error("Unable to record, file not provided")
		return 1
	}

	var fileToRecord string
	if terminal {
		fileToRecord = "terminal"
		app = true
	} else {
		fileToRecord = cmdFlags.Args()[0]
	}

	if app {
		fileToRecord = c.appToFile(fileToRecord)
	}

	if !(0 <= len(fileToRecord)) {
		return 0
	}

	if err := event.Record(fileToRecord); err != nil && !(err == project.ErrNotInitialized || err == project.ErrFileNotFound) {
		return 1
	} else if err == nil && status {
		var (
			err        error
			commitNote note.CommitNote
			out        string
			wd         string
		)

		wd, err = os.Getwd()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		defer func() {
			if err := os.Chdir(wd); err != nil {
				fmt.Printf("Unable to change back to working directory, %s\n", err)
			}
		}()

		err = os.Chdir(filepath.Dir(fileToRecord))
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		if commitNote, err = metric.Process(true); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		out, err = report.Status(commitNote, report.OutputOptions{TotalOnly: true, LongDuration: longDuration})
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		c.output(out)
	}

	return 0
}

// Given an app name creates (if it not was already created) the file ".gtm/{name}.app"
// that we use to track events, and returns the full path
func (c RecordCmd) appToFile(appName string) string {
	if !(len(appName) > 0) {
		return ""
	}
	projPath, err := scm.GitRepoPath()
	if err != nil {
		return ""
	}
	projPath, err = scm.Workdir(projPath)
	if err != nil {
		return ""
	}

	var file = filepath.Join(projPath, ".gtm", appName+".app")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		ioutil.WriteFile(
			file,
			[]byte{},
			0644)
	}

	return file
}

// Synopsis returns help
func (c RecordCmd) Synopsis() string {
	return "Record file and terminal events"
}
