// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/git-time-metric/gtm/command"
	"github.com/git-time-metric/gtm/project"
	"github.com/mitchellh/cli"
)

// Version is the released version set during the release build process
var Version = "0.0.0"

func main() {
	c := cli.NewCLI("gtm", Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init":   command.NewInit,
		"record": command.NewRecord,
		"commit": command.NewCommit,
		"report": command.NewReport,
		"status": command.NewStatus,
		"verify": command.NewVerify(Version),
		"uninit": command.NewUninit,
		"clean":  command.NewClean,
	}

	exitStatus, err := c.Run()
	if err != nil {
		if err := project.Log(err); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	os.Exit(exitStatus)
}
