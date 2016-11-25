package command

import (
	"os"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
	"github.com/mitchellh/cli"
)

func TestReportDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{""}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestReportAll(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{"-all"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}
func TestReportTimelineHours(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{"-format", "timeline-hours"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestReportTimelineCommits(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{"-format", "timeline-commits"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestReportFiles(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{"-format", "files"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestReportInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := ReportCmd{Ui: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm report(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm report(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
