package main

import (
	"fmt"
	"os"

	"edgeg.io/gtm/cmd"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("Git Time Metric", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"init":   cmd.NewInit,
		"record": cmd.NewRecord,
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(exitStatus)
}
