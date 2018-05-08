package event

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
)

var setup = func(t *testing.T) (util.TestRepo, func()) {
	x := util.NewTestRepo(t, false)
	f := func() { x.Remove() }
	x.Seed()
	util.CheckFatal(t, os.Chdir(x.PathIn("")), f)
	_, err := project.Initialize([]string{}, false)
	util.CheckFatal(t, err, f)
	return x, f
}

func TestNewApplicationFromName(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()

	a, err := NewApplicationFromName("Google Chrome")
	if err != nil {
		t.Errorf("want error nil got %s\n", err)
	}

	if a.Name() != "Google Chrome" {
		t.Errorf("want name test got %s\n", a.Name())
	}

	p := filepath.Join(repo.PathIn(""), ".gtm", "google-chrome.app")
	if a.Path() != p {
		t.Errorf("want path %s got %s\n", p, a.path)
	}

	if _, err := os.Stat(a.path); os.IsNotExist(err) {
		t.Errorf("want path %s to exist but got %s\n", a.path, err)
	}

	if !a.IsApplication() {
		t.Errorf("want IsApplication true got false")
	}

	if a.IsTerminal() {
		t.Error("want IsTerminal false got true")
	}
}

func TestNewApplicationFromPath(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()

	p := filepath.Join(repo.PathIn(""), ".gtm", "google-chrome.app")
	a := NewApplicationFromPath(p)
	if a.path != p {
		t.Errorf("want path %s got %s\n", p, a.path)
	}

	if a.name != "Google Chrome" {
		t.Errorf("want name 'Google Chrome' got %s\n", a.Name())
	}

}

func TestNewTerminalApplication(t *testing.T) {
	_, cleanup := setup(t)
	defer cleanup()
	x, err := NewTerminalApplication()
	if err != nil {
		t.Errorf("want error nil got %s", err)
	}
	if !x.IsTerminal() {
		t.Errorf("want IsTerminal true got false")
	}
}

func TestAppRecord(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()

	a, err := NewApplicationFromName("Google Chrome")
	if err != nil {
		t.Errorf("want error nil got %s\n", err)
	}

	saveNow := util.Now
	defer func() { util.Now = saveNow }()
	util.Now = func() time.Time { return time.Unix(1257894000, 0) }

	saveSetActive := project.SetActive
	defer func() { project.SetActive = saveSetActive }()

	project.SetActive = (func(path string) error {
		t.Errorf("want project.SetActive to not be called but was called with %s", path)
		return nil
	})

	a.Record()

	p := filepath.Join(repo.PathIn(""), ".gtm", "1257894000.event")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Errorf("want path %s to exist but got %s\n", p, err)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("want error nil got erro %s", err)
	}

	wantPath := filepath.Join(".gtm", "google-chrome.app")
	if string(b) != wantPath {
		t.Errorf("want %s for event file contents got %s", wantPath, string(b))
	}

}

func TestTerminalRecord(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()

	x, err := NewTerminalApplication()
	if err != nil {
		t.Errorf("want error nil got %s", err)
	}

	saveNow := util.Now
	defer func() { util.Now = saveNow }()
	util.Now = func() time.Time { return time.Unix(1257894000, 0) }

	saveSetActive := project.SetActive
	defer func() { project.SetActive = saveSetActive }()

	setActiveCalled := false
	project.SetActive = (func(path string) error {
		setActiveCalled = true
		return nil
	})

	x.Record()

	if !setActiveCalled {
		t.Errorf("project.SetActive was not called")
	}

	p := filepath.Join(repo.PathIn(""), ".gtm", "1257894000.event")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Errorf("want path %s to exist but got %s\n", p, err)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("want error nil got erro %s", err)
	}

	wantPath := filepath.Join(".gtm", "terminal.app")
	if string(b) != wantPath {
		t.Errorf("want %s for event file contents got %s", wantPath, string(b))
	}

}
