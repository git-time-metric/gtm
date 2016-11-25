package command

import (
	"os"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestUninitNotGTM(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	ui := new(cli.MockUi)
	c := UninitCmd{Ui: ui}

	args := []string{"-yes"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm uninit(%+v), want 1 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestUninitWithDefaults(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := UninitCmd{Ui: ui}

	args := []string{"-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm uninit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestUninitInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := UninitCmd{Ui: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm uninit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm uninit(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
