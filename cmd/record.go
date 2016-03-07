package cmd

import (
	"fmt"

	"edgeg.io/gtm/cfg"
	"edgeg.io/gtm/event"
	"edgeg.io/gtm/metric"

	"github.com/mitchellh/cli"
)

type RecordCmd struct {
}

func NewRecord() (cli.Command, error) {
	return RecordCmd{}, nil
}

func (r RecordCmd) Help() string {
	return `
	Record a file event

	gmetric record [file]

	The full path to the file is required when calling record.
	`
}

func (r RecordCmd) Run(args []string) int {
	if len(args) == 0 {
		fmt.Println("Unable to record, file not provided")
		return 1
	}

	//TODO: add an option to turn off silencing ErrFileNotFound errors
	gtmPath, err := event.Save(args[0])
	if err != nil && !(err == cfg.ErrNotInitialized || err == cfg.ErrFileNotFound) {
		fmt.Println(err)
		return 1
	}

	if err := metric.ProcessEvents(gtmPath); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

func (r RecordCmd) Synopsis() string {
	return `
	Record a file event
	`
}
