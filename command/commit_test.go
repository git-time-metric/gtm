package command

import (
	"os"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestCommitDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := CommitCmd{Ui: ui}

	args := []string{"-yes"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm commit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestCommitInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := CommitCmd{Ui: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm commit(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm commit(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
