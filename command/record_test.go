package command

import (
	"io/ioutil"
	"os"
	"os/exec"
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

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	(InitCmd{Ui: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RecordCmd{Ui: ui}

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

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	cmd := exec.Command("gtm", "init")
	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	ui := new(cli.MockUi)
	c := RecordCmd{Ui: ui}

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

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	cmd := exec.Command("gtm", "init")
	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	ui := new(cli.MockUi)
	c := RecordCmd{Ui: ui}

	args := []string{filepath.Join(repoPath, ".gtm", "README")}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	files, err := ioutil.ReadDir(filepath.Join(repoPath, ".gtm"))
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

	repoPath := repo.PathIn("")
	defer os.Chdir(repoPath)
	os.Chdir(repoPath)

	cmd := exec.Command("gtm", "init")
	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	ui := new(cli.MockUi)
	c := RecordCmd{Ui: ui}

	args := []string{"-terminal"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}

	files, err := ioutil.ReadDir(filepath.Join(repoPath, ".gtm"))
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
	c := RecordCmd{Ui: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	if rc != 1 {
		t.Errorf("gtm record(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm record(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
