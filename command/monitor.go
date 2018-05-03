package command

import (
	"errors"
	"flag"
	"log"
	"os"
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
Usage: gtm monitor [options]

  Record file or terminal events.

Options:

  -apps=""     list of applications to only monitor, i.e "safari,firefox,slack"
`
	return strings.TrimSpace(helpText)
}

// Run executes commit commands with args
func (c MonitorCmd) Run(args []string) int {
	var apps string
	cmdFlags := flag.NewFlagSet("monitor", flag.ContinueOnError)
	cmdFlags.StringVar(&apps, "apps", "", "")
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	applist := []string{}
	if strings.TrimSpace(apps) != "" {
		applist = strings.Split(apps, ",")
	}

	m := event.NewAppMonitor(
		func(app string) error {
			if (RecordCmd{Ui: &cli.ColoredUi{ErrorColor: cli.UiColorRed, Ui: &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin}}}).Run([]string{"-app", app}) > 1 {
				return errors.New("error recording")
			}
			return nil
		}, applist,
	)

	log.Print("starting application monitor")
	if err := m.Run(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	log.Print("stopping application monitor")
	return 0
}

// Synopsis return help for commit command
func (c MonitorCmd) Synopsis() string {
	return ""
}
