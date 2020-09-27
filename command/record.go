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

	"github.com/kilpkonn/gtm-enhanced/event"
	"github.com/kilpkonn/gtm-enhanced/metric"
	"github.com/kilpkonn/gtm-enhanced/note"
	"github.com/kilpkonn/gtm-enhanced/project"
	"github.com/kilpkonn/gtm-enhanced/report"
	"github.com/kilpkonn/gtm-enhanced/scm"

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
Usage: gtm record [options] [/path/file or app event]

  Record file or app events.

Options:

  -terminal=false            Record a terminal event.

  -status=false              Return total time recorded for event.

  -long-duration=false       Return total time recorded in long duration format.

  -app=false [event_name]    Record an app event.

  -run=false [app_name]		 Record run event.

  -build=false [app_name] 	 Record build event.
`
	return strings.TrimSpace(helpText)
}

// Run executes record command with args
func (c RecordCmd) Run(args []string) int {
	var status, terminal, run, build, longDuration, app bool
	var cwd string
	cmdFlags := flag.NewFlagSet("record", flag.ContinueOnError)
	cmdFlags.BoolVar(&status, "status", false, "")
	cmdFlags.BoolVar(&terminal, "terminal", false, "")
	cmdFlags.BoolVar(&run, "run", false, "")
	cmdFlags.BoolVar(&build, "build", false, "")
	cmdFlags.BoolVar(&longDuration, "long-duration", false, "")
	cmdFlags.BoolVar(&app, "app", false, "")
	cmdFlags.StringVar(&cwd, "cwd", "", "")
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
		fileToRecord = c.appToFile("terminal", cwd)
	} else if run {
		fileToRecord = c.runToFile(strings.ToLower(strings.Join(cmdFlags.Args(), "-")), cwd)
	} else if build {
		fileToRecord = c.buildToFile(strings.ToLower(strings.Join(cmdFlags.Args(), "-")), cwd)
	} else if app {
		fileToRecord = c.appToFile(strings.ToLower(strings.Join(cmdFlags.Args(), "-")), cwd) // TODO: list of configurable allowed options
	} else {
		fileToRecord = cmdFlags.Args()[0]
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
func (c RecordCmd) appToFile(appName string, cwd string) string {
	if len(cwd) <= 0 {
		return eventToFile(appName, "app")
	}
	return eventToFile(appName, "app", cwd)
}

func (c RecordCmd) runToFile(appName string, cwd string) string {
	if len(cwd) <= 0 {
		return eventToFile(appName, "run")
	}
	return eventToFile(appName, "run", cwd)
}

func (c RecordCmd) buildToFile(appName string, cwd string) string {
	if len(cwd) <= 0 {
		return eventToFile(appName, "build")
	}
	return eventToFile(appName, "build", cwd)
}

func eventToFile(event string, eventType string, cwd ...string) string {
	if !(len(event) > 0) {
		return ""
	}
	projPath, err := scm.GitRepoPath(cwd...)
	if err != nil {
		return ""
	}
	projPath, err = scm.Workdir(projPath)
	if err != nil {
		return ""
	}
	var file = filepath.Join(projPath, ".gtm", event+"."+eventType)
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
	return "Record file, terminal and app events"
}
