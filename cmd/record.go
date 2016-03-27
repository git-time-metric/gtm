package cmd

import (
	"fmt"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/event"

	"github.com/mitchellh/cli"
)

type RecordCmd struct {
}

func NewRecord() (cli.Command, error) {
	return RecordCmd{}, nil
}

func (r RecordCmd) Help() string {
	return `
	gtm record <full-path to a file>

	Records a timestamped file event that denotes when a file has been accessed 
	`
}

func (r RecordCmd) Run(args []string) int {
	if len(args) == 0 {
		fmt.Println("Unable to record, file not provided")
		return 1
	}

	//TODO: add an option to turn off silencing ErrFileNotFound errors
	if err := event.Record(args[0]); err != nil && !(err == env.ErrNotInitialized || err == env.ErrFileNotFound) {
		env.LogToGTM(err)
		return 1
	}

	return 0
}

func (r RecordCmd) Synopsis() string {
	return `
	Record a file event
	`
}
