package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

const (
	applicationExt = "app"
	gtmDirectory   = ".gtm"
	terminalName   = "Terminal"
)

type Application struct {
	name string
	path string
}

func NewApplicationFromName(name string) (Application, error) {
	a := Application{name: strings.TrimSpace(name)}
	if err := a.setFilePathFromName(); err != nil {
		return Application{}, err
	}
	if err := a.createFilePath(); err != nil {
		return Application{}, err
	}
	return a, nil
}

func NewTerminalApplication() (Application, error) {
	return NewApplicationFromName(terminalName)
}

func NewApplicationFromPath(path string) Application {
	a := Application{path: path}
	a.setNameFromFilePath()
	return a
}

func (a *Application) Record() error {
	p := project.GetActive()
	if p == "" {
		// if blank there is no currently active project
		return nil
	}

	x, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(x)

	err = os.Chdir(p)
	if err != nil {
		return nil
	}

	sourcePath, gtmPath, err := pathFromSource(a.path)
	if err != nil {
		return err
	}

	if err := writeEventFile(sourcePath, gtmPath); err != nil {
		return err
	}

	return nil
}

func (a *Application) setFilePathFromName() error {
	defer util.TimeTrack(time.Now(), "event.setFilePathFromName")

	projPath, err := scm.RootPath()
	if err != nil {
		return err
	}
	a.path = filepath.Join(projPath, gtmDirectory, fmt.Sprintf("%s.%s", normalizeAppName(a.name), applicationExt))
	return nil
}

func (a *Application) setNameFromFilePath() {
	n := filepath.Base(a.path)
	x := strings.LastIndex(n, fmt.Sprintf(".%s", applicationExt))
	if x > 0 {
		n = n[:x]
	}
	n = normalizedAppNameToTitle(n)
	n = strings.Title(n)
	a.name = n
}

func (a *Application) createFilePath() error {
	if _, err := os.Stat(a.path); os.IsNotExist(err) {
		if err := ioutil.WriteFile(a.path, []byte{}, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Name() string {
	return a.name
}

func (a *Application) Path() string {
	return a.path
}

func (a *Application) IsTerminal() bool {
	return a.name == terminalName
}

func (a *Application) IsApplication() bool {
	return strings.LastIndex(a.path, fmt.Sprintf(".%s", applicationExt)) > 0
}

func normalizeAppName(app string) string {
	return strings.ToLower(strings.Replace(app, " ", "-", -1))
}

func normalizedAppNameToTitle(app string) string {
	return strings.Title(strings.Replace(app, "-", " ", -1))
}
