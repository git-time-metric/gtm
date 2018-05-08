package event

import (
	"log"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/epoch"
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

	for {
		time.Sleep(time.Second * 5)

		var err error
		app, err = getFrontApp()
		if err != nil {
			return err
		}
		app = normalizeAppName(app)

		if app == prevApp && time.Unix(epoch.Now(), 0).Before(time.Unix(nextUpdate, 0)) {
			continue
		}
		prevApp = app
		nextUpdate = epoch.Now() + int64(updateInterval)

		if !m.Watching(app) {
			log.Printf("skipped %s\n", normalizedAppNameToTitle(app))
			continue
		}

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
