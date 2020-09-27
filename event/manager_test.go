// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package event

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kilpkonn/gtm-enhanced/project"
	"github.com/kilpkonn/gtm-enhanced/util"
)

func TestRecord(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	curDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(curDir)

	os.Chdir(repo.Workdir())

	repo.SaveFile("event.go", "event", "")
	sourceFile := filepath.Join(repo.Workdir(), "event", "event.go")
	if err = Record(sourceFile); err != project.ErrNotInitialized {
		t.Errorf("Record(%s), want error %s, got error %s", sourceFile, project.ErrNotInitialized, err)
	}

	project.Initialize(false, []string{}, false)

	sourceFile = filepath.Join(repo.Workdir(), "doesnotexist.go")
	if err = Record(sourceFile); err != project.ErrFileNotFound {
		t.Errorf("Record(%s), want error %s, got %s", sourceFile, project.ErrFileNotFound, err)
	}

	sourceFile = filepath.Join(repo.Workdir(), "event", "event.go")
	if err = Record(sourceFile); err != nil {
		t.Errorf("Record(%s), want error nil, got %s", sourceFile, err)
	}

	gtmPath := filepath.Join(repo.Workdir(), project.GTMDir)

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Errorf("Record(%s) returns error %s when reading .gtm directory", sourceFile, err)
	}
	if len(files) != 1 {
		t.Errorf("Record(%s), want file count 1, got %d", sourceFile, len(files))
	}

	b, err := ioutil.ReadFile(filepath.Join(gtmPath, files[0].Name()))
	if err != nil {
		t.Fatalf("Record(%s), unable to read event file %s, %s", sourceFile, files[0].Name(), err)
	}

	if !strings.Contains(string(b), filepath.Join("event", "event.go")) {
		t.Errorf("Record(%s), want file contents %s, got %s", sourceFile, filepath.Join("event", "event.go"), string(b))
	}
}

func TestProcess(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	curDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(curDir)

	os.Chdir(repo.Workdir())

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	expected := map[int64]map[string]int{
		int64(1458496800): {filepath.Join("event", "event.go"): 2, filepath.Join("event", "event_test.go"): 1},
		int64(1458496860): {filepath.Join("event", "event.go"): 1},
		int64(1458496920): {filepath.Join("event", "event.go"): 1},
	}

	workdir := repo.Workdir()
	gtmPath := filepath.Join(workdir, project.GTMDir)

	got, err := Process(gtmPath, true)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", workdir, gtmPath, err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Process(%s, %s, true)\nwant:\n%+v\ngot:\n%+v\n", workdir, gtmPath, expected, got)
	}

	got, err = Process(gtmPath, false)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", workdir, gtmPath, err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Process(%s, %s, true)\nwant:\n%+v\ngot:\n%+v", workdir, gtmPath, expected, got)
	}
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", workdir, gtmPath, err)
	}
	if len(files) != 0 {
		t.Fatalf("Process(%s, %s, true), want file count 0, got %d", workdir, gtmPath, len(files))
	}
}
