package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	d, err := ioutil.TempDir("", "gtm")
	if err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", d, err)
	}
	defer func() {
		if err = os.RemoveAll(d); err != nil {
			fmt.Printf("Error removing %s dir, %s", d, err)
		}
	}()

	savedCurDir, _ := os.Getwd()
	if err := os.Chdir(d); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	defer func() {
		if err = os.Chdir(savedCurDir); err != nil {
			fmt.Printf("Unable to change working directory, %s", err)
		}
	}()

	cmd := exec.Command("git", "init")
	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	s, err := Initialize()
	if err != nil {
		t.Errorf("Initialize(), want error nil got error %s", err)
	}
	if !strings.Contains(s, "Git Time Metric initialized") {
		t.Errorf("Initialize(), want Git Metric initialized, got %s", s)
	}

	fp := path.Join(d, ".git", "hooks", "post-commit")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Initialize(), want file post-commit, got %s", err)
	}
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Initialize(), want error nil, got %s", err)
	}
	if !strings.Contains(string(b), PostCommitHook+"\n") {
		t.Errorf("Initialize(), want %s got %s", PostCommitHook, string(b))
	}

	cmd = exec.Command("git", "config", "-l")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}
	if !strings.Contains(string(b), NotesRewriteRef) {
		t.Errorf("Initialize(), want %s got %s", NotesRewriteRef, string(b))
	}

	fp = path.Join(d, ".gitignore")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Initialize(), want file .gitignore, got %s", err)
	}
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Initialize(), want error nil, got %s", err)
	}
	if !strings.Contains(string(b), GitIgnore+"\n") {
		t.Errorf("Initialize(), want %s got %s", GitIgnore, string(b))
	}

}
