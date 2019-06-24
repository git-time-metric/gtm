// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestReportDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "2m 40s  89% [m] event/event.go"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportSummary(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-format", "summary", "-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "3m  0s This is a commit"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestProjectSummary(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-format", "project", "-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "3m  0s gtm"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportAll(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	// TODO: in order to test output of multi-project reporting, we need the ability to mock the project index
	args := []string{"-all"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestReportTimelineHours(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-format", "timeline-hours", "-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "Sun Mar 20"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportTimelineCommits(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-format", "timeline-commits", "-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "Wed Mar 06"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportFiles(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go"), filepath.Join("event", "event_test.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-format", "files", "-testing=true"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "3m  0s"
	if !strings.Contains(ui.OutputWriter.String(), want) {
		t.Errorf("gtm report(%+v), want %s got %s, %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportAppsOff(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("browser.app", project.GTMDir, "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join(project.GTMDir, "browser.app"))

	repo.Commit(repo.Stage(filepath.Join("event", "event.go")))

	// save notes to git repository
	(CommitCmd{UI: new(cli.MockUi)}).Run([]string{"-yes"})

	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	// Including apps
	args := []string{"-format", "files", "-testing=true"}
	rc := c.Run(args)
	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
	if !strings.Contains(ui.OutputWriter.String(), "Browser") {
		t.Errorf("gtm report(%+v), want 'Browser' got %s, %s", args, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}

	// Excluding apps
	ui.OutputWriter.Reset()
	ui.ErrorWriter.Reset()
	args = []string{"-app-off", "-format", "files", "-testing=true"}
	rc = c.Run(args)
	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
	if strings.Contains(ui.OutputWriter.String(), "Browser") {
		t.Errorf("gtm report(%+v), want not 'Browser' got %s, %s", args, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}

func TestReportInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := ReportCmd{UI: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm report(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
