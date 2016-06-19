package main

import (
	"fmt"
	"os"

	"github.com/git-time-metric/gtm/command"
	"github.com/git-time-metric/gtm/project"

	"github.com/mitchellh/cli"
)

var version string = "0.0.0"

func main() {
	c := cli.NewCLI("gtm", version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init":   command.NewInit,
		"record": command.NewRecord,
		"commit": command.NewCommit,
		"report": command.NewReport,
		"status": command.NewStatus,
	}

	exitStatus, err := c.Run()
	if err != nil {
		if err := project.Log(err); err != nil {
			fmt.Println(err)
		}
	}

	os.Exit(exitStatus)
}
