package command

import (
	"errors"
	"strings"

	"github.com/git-time-metric/gtm/event"
	"github.com/mitchellh/cli"
)

// MonitorCmd struct contain methods for monitor command
type MonitorCmd struct {
	Ui cli.Ui
}

// NewMonitor returns new MonitorCmd struct
func NewMonitor() (cli.Command, error) {
	return MonitorCmd{}, nil
}

// Help returns help for monitor command
func (c MonitorCmd) Help() string {
	helpText := `
Usage: gtm monitor
`
	return strings.TrimSpace(helpText)
}

// Run executes commit commands with args
func (c MonitorCmd) Run(args []string) int {
	m := event.NewAppMonitor(
		func(app string) error {
			if (RecordCmd{}).Run([]string{"-app", app}) > 1 {
				return errors.New("error recording")
			}
			return nil
		},
		[]string{"Google Chrome"},
	)

	if err := m.Run(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	return 0
}

// Synopsis return help for commit command
func (c MonitorCmd) Synopsis() string {
	return ""
}
