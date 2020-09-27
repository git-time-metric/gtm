// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"os"
	"strings"
	"testing"

	"github.com/kilpkonn/gtm-enhanced/util"
	"github.com/mitchellh/cli"
)

func TestCommitDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	os.Chdir(repo.Workdir())

	(InitCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := CommitCmd{UI: ui}

	args := []string{"-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm commit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestCommitInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := CommitCmd{UI: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm commit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm commit(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
