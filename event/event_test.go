package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/epoch"
)

func TestSave(t *testing.T) {

	// Setup directories and source file
	rootPath, err := ioutil.TempDir("", "gtm")
	if err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", rootPath, err)
	}
	defer func() {
		if err := os.RemoveAll(rootPath); err != nil {
			fmt.Printf("Error removing %s dir, %s", rootPath, err)
		}
	}()
	gtmPath := path.Join(rootPath, ".gtm")
	if err := os.MkdirAll(gtmPath, 0700); err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", gtmPath, err)
	}
	sourcePath := path.Join(rootPath, "src")
	if err := os.MkdirAll(sourcePath, 0700); err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", sourcePath, err)
	}
	sourceFile := path.Join(sourcePath, "source.go")
	if err := ioutil.WriteFile(sourceFile, []byte{}, 0600); err != nil {
		t.Fatalf("Unable to create tempory file %s, %s", sourceFile, err)
	}

	// Freeze the system time
	saveNow := env.Now
	env.Now = func() time.Time { return time.Unix(100, 0) }
	defer func() { env.Now = saveNow }()

	// Call save on an uninitialized Git Metric project
	if err := Save(sourceFile); err != env.ErrNotInitialized {
		t.Errorf("Save(%s), want error %s, got error %s", sourceFile, env.ErrNotInitialized, err)
	}

	// Replace env.Paths with a mock
	savePaths := env.Paths
	env.Paths = func(path ...string) (string, string, error) {
		return rootPath, gtmPath, nil
	}
	defer func() { env.Paths = savePaths }()

	// Call Save with an invalid source file
	if err := Save(path.Join(sourcePath, "doesnotexist.go")); err != env.ErrFileNotFound {
		t.Errorf("Save(%s), want error %s, got %s", sourceFile, env.ErrFileNotFound, err)
	}

	// Call Save with a valid source file
	if err := Save(sourceFile); err != nil {
		t.Errorf("Save(%s), want error nil, got %s", sourceFile, err)
	}

	// Is there one event file?
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Errorf("Save(%s) returns error %s when reading .gtm directory", sourceFile, err)
	}
	if len(files) != 1 {
		t.Errorf("Save(%s), want file count 1, got %d", sourceFile, len(files))
	}

	// Does the event file have the right prefix?
	if !strings.HasPrefix(files[0].Name(), fmt.Sprintf("%d-", epoch.MinuteNow())) {
		t.Errorf("Save(%s), want file prefix %s, got %s", sourceFile, fmt.Sprintf("%d-", epoch.MinuteNow()), files[0].Name())
	}

	// Read the event file
	b, err := ioutil.ReadFile(path.Join(gtmPath, files[0].Name()))
	eventContent := string(b)
	if err != nil {
		t.Fatalf("Save(%s), unable to read event file %s, %s", sourceFile, files[0].Name(), err)
	}

	// Does the event file have the right content?
	relPath, err := filepath.Rel(rootPath, sourceFile)
	if err != nil {
		t.Fatalf("Save(%s), error creating relative path for rootPath %s and sourcePath %s, %s", sourceFile, rootPath, sourcePath, err)
	}
	if fmt.Sprintf("%s,%s", rootPath, relPath) != strings.TrimSpace(string(eventContent)) {
		t.Errorf("Save(%s), want file contents %s, got %s", sourceFile, fmt.Sprintf("%s,%s", rootPath, relPath), string(eventContent))
	}
}

func TestSweep(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO: fix this, exec.Command("cp", "-r", fixturePath, rootPath) is not compatible with Windows
		fmt.Println("Skipping TestSweep, not compatible with Windows")
		return
	}

	expected := map[int64]map[string]int{
		int64(1458409500): map[string]int{"event/event.go": 1},
		int64(1458409560): map[string]int{"event/event.go": 2, "event/event_test.go": 2},
		int64(1458408780): map[string]int{"event/event_test.go": 1, "event/event.go": 1},
		int64(1458408840): map[string]int{"event/event_test.go": 1, "event/event.go": 1},
		int64(1458408900): map[string]int{"event/event_test.go": 1},
		int64(1458409020): map[string]int{"event/event_test.go": 1},
		int64(1458409080): map[string]int{"event/event_test.go": 1, "event/event.go": 1},
		int64(1458409140): map[string]int{"event/event.go": 1},
	}

	// Setup directories and copy fixtures
	rootPath, err := ioutil.TempDir("", "gtm")
	if err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", rootPath, err)
	}
	defer func() {
		if err := os.RemoveAll(rootPath); err != nil {
			fmt.Printf("Error removing %s dir, %s", rootPath, err)
		}
	}()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Sweep(), error getting current working directory, %s", err)
	}
	fixturePath := path.Join(wd, "test-fixtures")
	cmd := exec.Command("cp", "-r", fixturePath, rootPath)
	_, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to copy %s directory to %s", fixturePath, rootPath)
	}

	// sweep files with dry-run set to true
	gtmPath := path.Join(rootPath, "test-fixtures")
	actual, err := Sweep(gtmPath, true)
	if err != nil {
		t.Fatalf("Sweep(%s, true), want error nil, got %s", gtmPath, err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Sweep(%s, true), want %+v, got %+v", gtmPath, expected, actual)
	}

	// sweep files with dry-run set to false
	actual, err = Sweep(gtmPath, false)
	if err != nil {
		t.Fatalf("Sweep(%s, true), want error nil, got %s", gtmPath, err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Sweep(%s, true), want %+v, got %+v", gtmPath, expected, actual)
	}
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Fatalf("Sweep(%s, true), want error nil, got %s", gtmPath, err)
	}
	//NOTE: there is one txt file in the test fixtures that should not be removed
	if len(files) != 1 {
		t.Fatalf("Sweep(%s, true), want file count 0, got %d", gtmPath, len(files))
	}
}
