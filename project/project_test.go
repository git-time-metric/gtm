package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitialize(t *testing.T) {
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
		t.Errorf("Initialize(), want Git Time Metric has been initialized, got %s", s)
	}

	for hook, command := range GitHooks {
		fp := filepath.Join(d, ".git", "hooks", hook)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("Initialize(), want file post-commit, got %s", err)
		}
		if b, err = ioutil.ReadFile(fp); err != nil {
			t.Fatalf("Initialize(), want error nil, got %s", err)
		}
		if !strings.Contains(string(b), command+"\n") {
			t.Errorf("Initialize(), want %s got %s", command, string(b))
		}
	}

	cmd = exec.Command("git", "config", "-l")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}
	for k, v := range GitConfig {
		want := fmt.Sprintf("%s=%s", k, v)
		if !strings.Contains(string(b), want) {
			t.Errorf("Initialize(), want %s got %s", want, string(b))
		}
	}

	fp := filepath.Join(d, ".gitignore")
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

func TestUninitialize(t *testing.T) {
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
		t.Fatalf("Want error nil got error %s", err)
	}

	s, err = Uninitialize()
	if err != nil {
		t.Fatalf("Uninitialize(), want error nil got error %s", err)
	}
	if !strings.Contains(s, "Git Time Metric uninitialized") {
		t.Errorf("Uninitialize(), want Git Time Metric uninitialized, got %s", s)
	}

	for hook, command := range GitHooks {
		fp := filepath.Join(d, ".git", "hooks", hook)
		if b, err = ioutil.ReadFile(fp); err != nil {
			t.Fatalf("Uninitialize(), want error nil, got %s", err)
		}
		if strings.Contains(string(b), command+"\n") {
			t.Errorf("Uinitialize(), do not want %s got %s", command, string(b))
		}
	}

	cmd = exec.Command("git", "config", "-l")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Want error nil got error %s, %s", string(b), err)
	}
	for k, v := range GitConfig {
		donotwant := fmt.Sprintf("%s=%s", k, v)
		if strings.Contains(string(b), donotwant) {
			t.Errorf("Uninitialize(), do not want %s got %s", donotwant, string(b))
		}
	}

	fp := filepath.Join(d, ".gitignore")
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Uninitialize(), want error nil, got %s", err)
	}
	if strings.Contains(string(b), GitIgnore+"\n") {
		t.Errorf("Uninitialize(), do not want %s got %s", GitIgnore, string(b))
	}

	if _, err := os.Stat(path.Join(d, ".gtm")); !os.IsNotExist(err) {
		t.Errorf("Uninitialize(), error directory .gtm exists")
	}
}

func TestClean(t *testing.T) {
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

	_, err = Initialize()
	if err != nil {
		t.Fatalf("Want error nil got error %s", err)
	}

	gtmPath := filepath.Join(d, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		t.Fatalf("%s directory not found", gtmPath)
	}

	testFiles := []string{"a.event", "b.event", "a.metric", "a.txt"}
	for _, f := range testFiles {
		if err := ioutil.WriteFile(filepath.Join(gtmPath, f), []byte{}, 0644); err != nil {
			t.Errorf("Want error nil got %s", err)
		}
	}

	_, err = Clean()

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Want error nil got %s", err)
	}

	for _, f := range files {
		if f.Name() != "a.txt" {
			t.Errorf("Clean(), want only a.txt got %s", f.Name())
		}
	}
}
