// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"os"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestInitNoGitRepo(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	os.Chdir(repo.Workdir())
	repo.Remove()

	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm init(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestInitDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{}
	rc := c.Run(args)

	want := `
     post-commit: gtm commit --yes
  alias.fetchgtm: fetch origin refs/notes/gtm-data:refs/notes/gtm-data
   alias.pushgtm: push origin refs/notes/gtm-data
notes.rewriteref: refs/notes/gtm-data
        terminal: true
      .gitignore: /.gtm/
            tags:
`
	if rc != 0 {
		t.Errorf("gtm init(%+v), want 0 got %d", args, rc)
	}
	if !strings.Contains(strings.TrimSpace(ui.OutputWriter.String()), strings.TrimSpace(want)) {
		t.Errorf("gtm init(%+v), want %s got %s", args, want, ui.OutputWriter.String())
	}
}

func TestInitTerminalFalse(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{"-terminal=false"}
	rc := c.Run(args)

	want := "terminal: false"

	if rc != 0 {
		t.Errorf("gtm init(%+v), want 0 got %d", args, rc)
	}
	if !strings.Contains(strings.TrimSpace(ui.OutputWriter.String()), strings.TrimSpace(want)) {
		t.Errorf("gtm init(%+v), want %s got %s", args, want, ui.OutputWriter.String())
	}
}

func TestInitTags(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{"-tags=t1,t2"}
	rc := c.Run(args)

	want := "tags: t1 t2"

	if rc != 0 {
		t.Errorf("gtm init(%+v), want 0 got %d", args, rc)
	}
	if !strings.Contains(strings.TrimSpace(ui.OutputWriter.String()), strings.TrimSpace(want)) {
		t.Errorf("gtm init(%+v), want %s got %s", args, want, ui.OutputWriter.String())
	}
}

func TestClearTags(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{"-clear-tags"}
	rc := c.Run(args)

	want := "tags:    "

	if rc != 0 {
		t.Errorf("gtm init(%+v), want 0 got %d", args, rc)
	}
	if !strings.Contains(strings.TrimSpace(ui.OutputWriter.String()), strings.TrimSpace(want)) {
		t.Errorf("gtm init(%+v), want %s got %s", args, want, ui.OutputWriter.String())
	}
}

func TestInitInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := InitCmd{Ui: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm init(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm init(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
