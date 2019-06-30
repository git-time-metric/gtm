// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestRecordInvalidFile(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{"nofile.txt"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestRecordNoFile(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{""}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestRecordFile(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{filepath.Join(workdir, "README")}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordFileWithStatus(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui, Out: new(bytes.Buffer)}

	args := []string{"-status", filepath.Join(workdir, "README")}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	if c.Out.String() != "1m0s" {
		t.Errorf("gtm record(%+v), want '1m0s' got %s", args, c.Out.String())
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordFileWithStatusLongDuration(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui, Out: new(bytes.Buffer)}

	args := []string{"-status", "-long-duration", filepath.Join(workdir, "README")}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	if c.Out.String() != "1 minute" {
		t.Errorf("gtm record(%+v), want '1 minutes' got %s", args, c.Out.String())
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordTerminal(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{"-terminal"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordTerminalWithStatus(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui, Out: new(bytes.Buffer)}

	args := []string{"-terminal", "-status"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	if c.Out.String() != "1m0s" {
		t.Errorf("gtm record(%+v), want '1m0s' got %s", args, c.Out.String())
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordApp(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	workdir := repo.Workdir()
	os.Chdir(workdir)

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{"-app", "browser"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	if _, err := os.Stat(filepath.Join(workdir, ".gtm", "browser.app")); os.IsNotExist(err) {
		t.Errorf("gtm record(%+v), want .app file to be created, it was not created", args)
	}

	files, err := ioutil.ReadDir(filepath.Join(workdir, ".gtm"))
	if err != nil {
		t.Fatalf("gtm record(%+v), want error nil got  %s", args, err)
	}
	cnt := 1
	for _, f := range files {
		if filepath.Base(f.Name()) == ".event" {
			cnt++
		}
	}
	if cnt != 1 {
		t.Errorf("gtm record(%+v), want 1 event file got %d, %s", args, cnt, ui.ErrorWriter.String())
	}
}

func TestRecordInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := RecordCmd{UI: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm record(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
