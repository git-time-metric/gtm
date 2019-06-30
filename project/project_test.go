// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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

	"github.com/git-time-metric/gtm/util"
)

func TestInitialize(t *testing.T) {
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

	s, err := Initialize(false, []string{}, false)
	if err != nil {
		t.Errorf("Initialize(), want error nil got error %s", err)
	}
	if !strings.Contains(s, "Git Time Metric initialized") {
		t.Errorf("Initialize(), want Git Time Metric has been initialized, got %s", s)
	}

	for hook, command := range GitHooks {
		fp := filepath.Join(rootPath, ".git", "hooks", hook)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("Initialize(), want file post-commit, got %s", err)
		}
		if b, err = ioutil.ReadFile(fp); err != nil {
			t.Fatalf("Initialize(), want error nil, got %s", err)
		}
		if !strings.Contains(string(b), command.Command) {
			t.Errorf("Initialize(), want %s got %s", command.Command, string(b))
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

	fp := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Initialize(), want file .gitignore, got %s", err)
	}
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Initialize(), want error nil, got %s", err)
	}
	if !strings.Contains(string(b), GitIgnore+"\n") {
		t.Errorf("Initialize(), want %s got %s", GitIgnore, string(b))
	}
	fp = filepath.Join(rootPath, ".gtm", "terminal.app")
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		t.Errorf("Initialize(), want file terminal.app does not exist, got %s", err)
	}

	// let's reinitialize with terminal tracking enabled
	s, err = Initialize(true, []string{}, false)
	if err != nil {
		t.Errorf("Initialize(true), want error nil got error %s", err)
	}
	if !strings.Contains(s, "Git Time Metric initialized") {
		t.Errorf("Initialize(true), want Git Time Metric has been initialized, got %s", s)
	}

	for hook, command := range GitHooks {
		fp := filepath.Join(rootPath, ".git", "hooks", hook)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("Initialize(true), want file post-commit, got %s", err)
		}
		if b, err = ioutil.ReadFile(fp); err != nil {
			t.Fatalf("Initialize(true), want error nil, got %s", err)
		}
		if !strings.Contains(string(b), command.Command) {
			t.Errorf("Initialize(true), want %s got %s", command.Command, string(b))
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
			t.Errorf("Initialize(true), want %s got %s", want, string(b))
		}
	}

	fp = filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Initialize(true), want file .gitignore, got %s", err)
	}
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Initialize(true), want error nil, got %s", err)
	}
	if !strings.Contains(string(b), GitIgnore+"\n") {
		t.Errorf("Initialize(true), want %s got %s", GitIgnore, string(b))
	}
	fp = filepath.Join(rootPath, ".gtm", "terminal.app")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Initialize(true), want file terminal.app, got %s", err)
	}
}

func TestUninitialize(t *testing.T) {
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

	_, err = Initialize(false, []string{}, false)
	if err != nil {
		t.Fatalf("Want error nil got error %s", err)
	}

	s, err := Uninitialize()
	if err != nil {
		t.Fatalf("Uninitialize(), want error nil got error %s", err)
	}
	if !strings.Contains(s, "Git Time Metric uninitialized") {
		t.Errorf("Uninitialize(), want Git Time Metric uninitialized, got %s", s)
	}

	for hook, command := range GitHooks {
		fp := filepath.Join(rootPath, ".git", "hooks", hook)
		if b, err = ioutil.ReadFile(fp); err != nil {
			t.Fatalf("Uninitialize(), want error nil, got %s", err)
		}
		if strings.Contains(string(b), command.Command+"\n") {
			t.Errorf("Uinitialize(), do not want %s got %s", command.Command, string(b))
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

	fp := filepath.Join(rootPath, ".gitignore")
	if b, err = ioutil.ReadFile(fp); err != nil {
		t.Fatalf("Uninitialize(), want error nil, got %s", err)
	}
	if strings.Contains(string(b), GitIgnore+"\n") {
		t.Errorf("Uninitialize(), do not want %s got %s", GitIgnore, string(b))
	}

	if _, err := os.Stat(path.Join(rootPath, ".gtm")); !os.IsNotExist(err) {
		t.Errorf("Uninitialize(), error directory .gtm exists")
	}
}

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

	_, err = Initialize(false, []string{}, false)
	if err != nil {
		t.Fatalf("Want error nil got error %s", err)
	}

	gtmPath := filepath.Join(rootPath, GTMDir)
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

	// write an app event file
	if err := ioutil.WriteFile(filepath.Join(gtmPath, "app.event"), []byte(".gtm/browser.app"), 0644); err != nil {
		t.Errorf("Want error nil got %s", err)
	}

	// lets only delete terminal events
	err = Clean(util.AfterNow(0), true, false)
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Want error nil got %s", err)
	}
	for _, f := range files {
		if !(f.Name() == "a.txt" || f.Name() == "a.event" || f.Name() == "b.event" || f.Name() == "app.event") {
			t.Errorf("Clean(), want only a.txt, a.event, b.event and app.event got %s", f.Name())
		}
	}

	// lets only delete app events
	err = Clean(util.AfterNow(0), false, true)
	files, err = ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Want error nil got %s", err)
	}
	for _, f := range files {
		if !(f.Name() == "a.txt" || f.Name() == "a.event" || f.Name() == "b.event") {
			t.Errorf("Clean(), want only a.txt, a.event and b.event got %s", f.Name())
		}
	}

	// lets clean all events
	err = Clean(util.AfterNow(0), false, false)
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
