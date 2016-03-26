package main

import (
	"fmt"
	"os"

	"edgeg.io/gtm/cmd"
	"edgeg.io/gtm/env"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("Git Time Metric", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init":   cmd.NewInit,
		"record": cmd.NewRecord,
		"commit": cmd.NewCommit,
	}

	exitStatus, err := c.Run()
	if err != nil {
		if err := env.LogToGTM(err); err != nil {
			fmt.Println(err)
		}
	}

	os.Exit(exitStatus)
}
