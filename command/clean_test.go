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

func TestCleanYes(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := CleanCmd{UI: ui}

	args := []string{"-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm clean(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestTerminalOnly(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := CleanCmd{UI: ui}

	args := []string{"-terminal-only", "-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm clean(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestAppOnly(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	repo.SaveFile("browser.app", project.GTMDir, "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458497804.event", project.GTMDir, filepath.Join(project.GTMDir, "browser.app"))

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := CleanCmd{UI: ui}

	args := []string{"-app-only", "-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm clean(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	if !repo.FileExists("1458496803.event", project.GTMDir) {
		t.Errorf("gtm clean(%+v), want non-app event to not be deleted, but was deleted", args)
	}

	if repo.FileExists("1458497804.event", project.GTMDir) {
		t.Errorf("gtm clean(%+v), want app event to be deleted, but was found", args)
	}
}

func TestCleanInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := CleanCmd{UI: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm clean(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm clean(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
