package main

import (
	"fmt"
	"os"

	"edgeg.io/gtm/command"
	"edgeg.io/gtm/project"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("Git Time Metric", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init":   command.NewInit,
		"record": command.NewRecord,
		"commit": command.NewCommit,
		"report": command.NewReport,
	}

	exitStatus, err := c.Run()
	if err != nil {
		if err := project.Log(err); err != nil {
			fmt.Println(err)
		}
	}

	os.Exit(exitStatus)
}
