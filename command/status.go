package command

import (
	"fmt"

	"edgeg.io/gtm/metric"
	"edgeg.io/gtm/report"
	"github.com/mitchellh/cli"
)

type StatusCmd struct {
}

func NewStatus() (cli.Command, error) {
	return StatusCmd{}, nil
}

func (r StatusCmd) Help() string {
	return `
	`
}

func (r StatusCmd) Run(args []string) int {
	out := ""
	if commitNote, err := metric.Process(true); err != nil {
		fmt.Println(err)
		return 1
	} else {
		out, err = report.NoteFiles(commitNote)
		if err != nil {
			fmt.Println(err)
			return 1
		}
	}
	fmt.Printf(out)
	return 0
}

func (r StatusCmd) Synopsis() string {
	return `
	`
}
