package event

import (
	"log"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/epoch"
	"github.com/git-time-metric/robotgo"
	ps "github.com/mitchellh/go-ps"
)

var (
	updateInterval = time.Second * 30
)

type AppMonitor struct {
	RecordFunc   func(app string) error
	Applications map[string]bool
}

func NewAppMonitor(recordFunc func(app string) error, appsToMonitor []string) AppMonitor {
	apps := map[string]bool{}
	for _, x := range appsToMonitor {
		apps[normalizeAppName(strings.TrimSpace(x))] = true
	}
	return AppMonitor{RecordFunc: recordFunc, Applications: apps}
}

func (m *AppMonitor) Run() error {
	var (
		app, prevApp string
		nextUpdate   = epoch.Now()
	)

	// cmd, err := exec.LookPath("osascript")
	// if err != nil {
	// 	return err
	// }

	for {
		time.Sleep(time.Second * 1)

		// x := exec.Command(
		// 	cmd,
		// 	`-e`, `tell application "System Events"`,
		// 	`-e`, `set frontApp to name of first application process whose frontmost is true`,
		// 	`-e`, `end tell`,
		// )

		// o, err := x.CombinedOutput()
		// if err != nil {
		// 	return err
		// }

		// app = normalizeAppName(strings.Replace(string(o), "\n", "", -1))

		pid := robotgo.GetPID()
		x, err := ps.FindProcess(pid)
		if err != nil || x == nil {
			log.Printf("error finding process for pid %d\n", pid)
			continue
		}
		app = normalizeAppName(x.Executable())

		if app == prevApp && time.Unix(epoch.Now(), 0).Before(time.Unix(nextUpdate, 0)) {
			continue
		}
		prevApp = app
		nextUpdate = epoch.Now() + int64(updateInterval)

		if !m.Watching(app) {
			log.Printf("skipped %s\n", normalizedAppNameToTitle(app))
			continue
		}

		log.Printf("watching process id %d application %s\n", pid, app)

		if err := m.RecordFunc(app); err != nil {
			return err
		}
	}
}

func (m *AppMonitor) Watching(app string) bool {
	if len(m.Applications) == 0 {
		return true
	}
	return m.Applications[app]
}
