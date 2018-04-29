package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
)

func TestClean(t *testing.T) {
	rootPath, err := ioutil.TempDir("", "gtm")
	if err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", rootPath, err)
	}
	defer func() {
		if err = os.RemoveAll(rootPath); err != nil {
			fmt.Printf("Error removing %s dir, %s", rootPath, err)
		}
	}()

	savedCurDir, _ := os.Getwd()
	if err := os.Chdir(rootPath); err != nil {
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

	_, err = project.Initialize([]string{}, false)
	if err != nil {
		t.Fatalf("Want error nil got error %s", err)
	}

	gtmPath := filepath.Join(rootPath, project.GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		t.Fatalf("%s directory not found", gtmPath)
	}

	testFiles := []string{"a.event", "b.event", "a.metric", "a.txt"}
	for _, f := range testFiles {
		if err := ioutil.WriteFile(filepath.Join(gtmPath, f), []byte{}, 0644); err != nil {
			t.Errorf("Want error nil got %s", err)
		}
	}
	// write a terminal event file
	if err := ioutil.WriteFile(filepath.Join(gtmPath, "t.event"), []byte("terminal.app"), 0644); err != nil {
		t.Errorf("Want error nil got %s", err)
	}

	// lets only delete terminal events
	err = Clean(util.AfterNow(0), false, false, true)
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Want error nil got %s", err)
	}
	for _, f := range files {
		if !(f.Name() == "a.txt" || f.Name() == "a.event" || f.Name() == "b.event") {
			t.Errorf("Clean(), want only a.txt, a.event and b.event got %s", f.Name())
		}
	}

	// lets clean all events
	err = Clean(util.AfterNow(0), true, true, true)
	files, err = ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Want error nil got %s", err)
	}
	for _, f := range files {
		if f.Name() != "a.txt" {
			t.Errorf("Clean(), want only a.txt got %s", f.Name())
		}
	}
}
