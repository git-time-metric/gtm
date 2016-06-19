package event

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
)

func TestRecord(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	curDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(curDir)

	os.Chdir(repo.PathIn(""))

	repo.SaveFile("event.go", "event", "")
	sourceFile := filepath.Join(repo.PathIn(""), "event", "event.go")
	if err = Record(sourceFile); err != project.ErrNotInitialized {
		t.Errorf("Record(%s), want error %s, got error %s", sourceFile, project.ErrNotInitialized, err)
	}

	project.Initialize()

	sourceFile = filepath.Join(repo.PathIn(""), "doesnotexist.go")
	if err = Record(sourceFile); err != project.ErrFileNotFound {
		t.Errorf("Record(%s), want error %s, got %s", sourceFile, project.ErrFileNotFound, err)
	}

	sourceFile = filepath.Join(repo.PathIn(""), "event", "event.go")
	if err = Record(sourceFile); err != nil {
		t.Errorf("Record(%s), want error nil, got %s", sourceFile, err)
	}

	gtmPath := filepath.Join(repo.PathIn(""), project.GTMDir)

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

	os.Chdir(repo.PathIn(""))

	repo.SaveFile("event.go", "event", "")
	repo.SaveFile("event_test.go", "event", "")
	repo.SaveFile("1458496803.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496811.event", project.GTMDir, filepath.Join("event", "event_test.go"))
	repo.SaveFile("1458496818.event", project.GTMDir, filepath.Join("event", "event.go"))
	repo.SaveFile("1458496943.event", project.GTMDir, filepath.Join("event", "event.go"))

	// NOTE - last two are idle events, 1458496980 & 1458497040
	expected := map[int64]map[string]int{
		int64(1458496800): map[string]int{filepath.Join("event", "event.go"): 2, filepath.Join("event", "event_test.go"): 1},
		int64(1458496860): map[string]int{filepath.Join("event", "event.go"): 1},
		int64(1458496920): map[string]int{filepath.Join("event", "event.go"): 1},
		int64(1458496980): map[string]int{filepath.Join("event", "event.go"): 1},
		int64(1458497040): map[string]int{filepath.Join("event", "event.go"): 1},
	}

	rootPath := repo.PathIn("")
	gtmPath := filepath.Join(rootPath, project.GTMDir)

	got, err := Process(rootPath, gtmPath, true)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", rootPath, gtmPath, err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Process(%s, %s, true)\nwant:\n%+v\ngot:\n%+v\n", rootPath, gtmPath, expected, got)
	}

	got, err = Process(rootPath, gtmPath, false)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", rootPath, gtmPath, err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Process(%s, %s, true)\nwant:\n%+v\ngot:\n%+v", rootPath, gtmPath, expected, got)
	}
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Process(%s, %s, true), want error nil, got %s", rootPath, gtmPath, err)
	}
	if len(files) != 0 {
		t.Fatalf("Process(%s, %s, true), want file count 0, got %d", rootPath, gtmPath, len(files))
	}
}
