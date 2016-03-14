package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"edgeg.io/gtm/env"
	"edgeg.io/gtm/epoch"
)

func TestSave(t *testing.T) {

	// Create directories and source file
	rootPath, err := ioutil.TempDir("", "gtm")
	if err != nil {
		t.Errorf("Unable to create tempory directory %s for TestSave, %s", rootPath, err)
	}
	gtmPath := path.Join(rootPath, ".gtm")
	if err := os.MkdirAll(gtmPath, 0700); err != nil {
		t.Errorf("Unable to create tempory directory %s for TestSave, %s", gtmPath, err)
	}
	sourcePath := path.Join(rootPath, "src")
	if err := os.MkdirAll(sourcePath, 0700); err != nil {
		t.Errorf("Unable to create tempory directory %s for TestSave, %s", sourcePath, err)
	}
	sourceFile := path.Join(sourcePath, "source.go")
	if err := ioutil.WriteFile(sourceFile, []byte{}, 0600); err != nil {
		t.Errorf("Unable to create tempory file %s for TestSave, %s", sourceFile, err)
	}
	defer os.Remove(sourceFile)
	defer os.Remove(rootPath)

	// Freeze the system time
	env.SetNow(time.Now())
	defer env.ClearNow()

	// Replace env.Paths with a mock
	savePaths := env.Paths
	env.Paths = func(path ...string) (string, string, error) {
		return rootPath, gtmPath, nil
	}
	defer func() { env.Paths = savePaths }()

	// Call Save with source file
	if err := Save(sourceFile); err != nil {
		t.Errorf("Save(%s) returns error %s", sourceFile, err)
	}

	// Is there one event file?
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		t.Errorf("Save(%s) returns error %s when reading .gtm directory", sourceFile, err)
	}
	if len(files) != 1 {
		t.Errorf("Save(%s), want event file count 1 but got %d", sourceFile, len(files))
	}

	// Does the event file have the right prefix?
	if !strings.HasPrefix(files[0].Name(), fmt.Sprintf("%d-", epoch.MinuteNow())) {
		t.Errorf("Save(%s), want event file prefix %s but got %s", sourceFile, fmt.Sprintf("%d-", epoch.MinuteNow()), files[0].Name())
	}

	// Read the event file
	b, err := ioutil.ReadFile(path.Join(gtmPath, files[0].Name()))
	eventContent := string(b)
	if err != nil {
		t.Errorf("Save(%s), unable to read event file %s, %s", sourceFile, files[0].Name(), err)
	}

	// Does the event file have the right content?
	relPath, err := filepath.Rel(rootPath, sourceFile)
	if err != nil {
		t.Errorf("Save(%s), unable to create relative path for rootPath %s and sourcePath %s, %s", sourceFile, rootPath, sourcePath, err)
	}
	if fmt.Sprintf("%s,%s", rootPath, relPath) != strings.TrimSpace(string(eventContent)) {
		t.Errorf("Save(%s), want file contents %s but got %s", sourceFile, fmt.Sprintf("%s,%s", rootPath, relPath), string(eventContent))
	}
}
