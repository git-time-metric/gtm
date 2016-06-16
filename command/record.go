package command

import (
	"fmt"

	"edgeg.io/gtm/event"
	"edgeg.io/gtm/project"

	"github.com/mitchellh/cli"
)

type RecordCmd struct {
}

func NewRecord() (cli.Command, error) {
	return RecordCmd{}, nil
}

func (r RecordCmd) Help() string {
	return r.Synopsis()
}

func (r RecordCmd) Run(args []string) int {
	if len(args) == 0 {
		fmt.Println("Unable to record, file not provided")
		return 1
	}

	//TODO: add an option to turn off silencing ErrFileNotFound errors
	if err := event.Record(args[0]); err != nil && !(err == project.ErrNotInitialized || err == project.ErrFileNotFound) {
		if err := project.Log(err); err != nil {
			fmt.Println(err)
		}
		return 1
	}

	return 0
}

func (r RecordCmd) Synopsis() string {
	return `
	Usage: gtm record <filepath>
	Record a file event
	`
}
