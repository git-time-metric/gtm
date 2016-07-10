package main

import (
	"fmt"
	"os"

	"github.com/git-time-metric/gtm/command"
	"github.com/git-time-metric/gtm/project"
	"github.com/mitchellh/cli"
)

var Version string = "0.0.0"

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
	}

	exitStatus, err := c.Run()
	if err != nil {
		if err := project.Log(err); err != nil {
			fmt.Println(err)
		}
	}

	os.Exit(exitStatus)
}
