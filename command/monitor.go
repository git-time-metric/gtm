package command

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/epoch"
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

// TODO: name and regex
// i.e.  "Research": `.*googelchrome.*|.*safari.*|.*firefox.*`

var (
	appFilter = map[string]bool{
		"googlechrome": true,
		"safari":       true,
		"slack":        true,
		"firefox":      true,
	}
)

// Run executes commit commands with args
func (c MonitorCmd) Run(args []string) int {

	cmd, err := exec.LookPath("osascript")
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	ui := &cli.ColoredUi{ErrorColor: cli.UiColorRed, Ui: &cli.BasicUi{Writer: os.Stdout, Reader: os.Stdin}}
	record := RecordCmd{Ui: ui}

	var app, prevApp string
	var nextUpdate int64 = epoch.Now()

	for {
		time.Sleep(time.Second * 1)

		x := exec.Command(cmd, `-e`, `tell application "System Events"`, `-e`, `set frontApp to name of first application process whose frontmost is true`, `-e`, `end tell`)
		x.Stdin = strings.NewReader("some input")
		var out bytes.Buffer
		x.Stdout = &out
		err := x.Run()
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}

		app = strings.Replace(out.String(), "\n", "", -1)

		if app == prevApp && time.Unix(epoch.Now(), 0).Before(time.Unix(nextUpdate, 0)) {
			continue
		}
		prevApp = app
		nextUpdate = epoch.Now() + 30

		if _, found := appFilter[normalize(app)]; !found {
			log.Printf("skipped %s\n", app)
			continue
		}

		log.Printf("recording %s\n", app)
		if record.Run([]string{"-app", app}) > 1 {
			c.Ui.Error(err.Error())
			return 1
		}
	}

	return 0
}

func normalize(app string) string {
	return strings.ToLower(strings.Replace(app, " ", "", -1))
}

// Synopsis return help for commit command
func (c MonitorCmd) Synopsis() string {
	return ""
}
