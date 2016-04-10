package metric

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"testing"

	"edgeg.io/gtm/env"
)

func TestProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO: fix this, exec.Command("cp", path.Join(fixturePath, f.Name()), gtmPath) is not compatible with Windows
		fmt.Println("Skipping TestSweep, not compatible with Windows")
		return
	}

	rootPath, _, f1 := processSetup(t)
	defer f1()

	var (
		cmd *exec.Cmd
	)

	// Test process with committing both git tracked files that have been modified

	// chandge working directory and initialize git repo
	savedCurDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(savedCurDir); err != nil {
			fmt.Printf("Unable to change working directory, %s", err)
		}
	}()
	if err := os.Chdir(rootPath); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	cmd = exec.Command("git", "init")
	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	// commit source files to git repo
	cmd = exec.Command("git", "add", "event/")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git add, %s", string(b))
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git commit, %s", string(b))
	}

	err = Process(false, false)
	if err != nil {
		t.Fatalf("Process(false, false) - test full commit, want error nil, got %s", err)
	}

	cmd = exec.Command("git", "notes", "--ref", env.NoteNameSpace, "show")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git notes, %s", string(b))
	}

	want := []string{`total:300.*`, `event.go:280.*,m`, `event_test.go:20.*,m`}
	for _, s := range want {
		matched, err := regexp.MatchString(s, string(b))
		if err != nil {
			t.Fatalf("Unable to run regexp.MatchString(%s, %s), %s", s, string(b), err)
		}
		if !matched {
			t.Errorf("Process(false, false) - test full commit, \nwant:\n%s,\ngot:\n%s", s, string(b))
		}
	}

	// Test Process by committing a tracked file that has been modified and one untracked file that is not added/commited

	// change back to saved current working directory and setup
	if err := os.Chdir(savedCurDir); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	rootPath, gtmPath, f2 := processSetup(t)
	defer f2()

	// chandge working directory and initialize git repo
	if err := os.Chdir(rootPath); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	cmd = exec.Command("git", "init")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	// commit source files to git repo
	cmd = exec.Command("git", "add", "event/event_test.go")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git add, %s", string(b))
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git commit, %s", string(b))
	}

	err = Process(false, false)
	if err != nil {
		t.Fatalf("Process(false, false), want error nil, got %s", err)
	}

	cmd = exec.Command("git", "notes", "--ref", env.NoteNameSpace, "show")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git notes, %s", string(b))
	}

	want = []string{`total:20`, `event_test.go:20.*,m`}
	for _, s := range want {
		matched, err := regexp.MatchString(s, string(b))
		if err != nil {
			t.Fatalf("Unable to run regexp.MatchString(%s, %s), %s", s, string(b), err)
		}
		if !matched {
			t.Errorf("Process(false, false) - test partial commit, \nwant:\n%s,\ngot:\n%s", s, string(b))
		}

	}
	fp := path.Join(gtmPath, "6f53bc90ba625b5afaac80b422b44f1f609d6367.metric")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Process(false, false) - test partial commit, want file %s exist, got file exists false", fp)
	}

	// Test Process by committing a tracked file that has been modified and one tracked file that is unmodified

	// change back to saved current working directory and setup
	if err := os.Chdir(savedCurDir); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	rootPath, gtmPath, f3 := processSetup(t)
	defer f3()

	// chandge working directory and initialize git repo
	if err := os.Chdir(rootPath); err != nil {
		t.Fatalf("Unable to change working directory, %s", err)
	}
	cmd = exec.Command("git", "init")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to initialize git repo, %s", string(b))
	}

	// commit source files to git repo
	cmd = exec.Command("git", "add", "event/event_test.go")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git add, %s", string(b))
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git commit, %s", string(b))
	}

	// commit source files to git repo
	cmd = exec.Command("git", "add", "event/event.go")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git add, %s", string(b))
	}
	cmd = exec.Command("git", "commit", "-m", "Second commit")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git commit, %s", string(b))
	}

	err = Process(false, false)
	if err != nil {
		t.Fatalf("Process(false, false) - test commit with readonly, want error nil, got %s", err)
	}

	cmd = exec.Command("git", "notes", "--ref", env.NoteNameSpace, "show")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git notes, %s", string(b))
	}

	want = []string{`total:300`, `event_test.go:20.*,r`, `event/event.go:280.*,m`}
	for _, s := range want {
		matched, err := regexp.MatchString(s, string(b))
		if err != nil {
			t.Fatalf("Unable to run regexp.MatchString(%s, %s), %s", s, string(b), err)
		}
		if !matched {
			t.Errorf("Process(false, false) - test commit with readonly, \nwant:\n%s,\ngot:\n%s", s, string(b))
		}

	}
}

func processSetup(t *testing.T) (string, string, func()) {
	var (
		rootPath   string
		gtmPath    string
		sourcePath string
		sourceFile string
		err        error
	)

	// setup directories and source files
	rootPath, err = ioutil.TempDir("", "gtm")
	if err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", rootPath, err)
	}
	gtmPath = path.Join(rootPath, env.GTMDirectory)
	if err = os.MkdirAll(gtmPath, 0700); err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", gtmPath, err)
	}
	sourcePath = path.Join(rootPath, "event")
	if err = os.MkdirAll(sourcePath, 0700); err != nil {
		t.Fatalf("Unable to create tempory directory %s, %s", sourcePath, err)
	}
	sourceFile = path.Join(sourcePath, "event.go")
	if err = ioutil.WriteFile(sourceFile, []byte{}, 0600); err != nil {
		t.Fatalf("Unable to create tempory file %s, %s", sourceFile, err)
	}
	sourceFile = path.Join(sourcePath, "event_test.go")
	if err = ioutil.WriteFile(sourceFile, []byte{}, 0600); err != nil {
		t.Fatalf("Unable to create tempory file %s, %s", sourceFile, err)
	}

	// replace env.Paths with a mock
	savePaths := env.Paths
	env.Paths = func(path ...string) (string, string, error) {
		return rootPath, gtmPath, nil
	}

	var (
		wd          string
		fixturePath string
		cmd         *exec.Cmd
		files       []os.FileInfo
	)

	// copy fixtures
	wd, err = os.Getwd()
	if err != nil {
		t.Fatalf("Sweep(), error getting current working directory, %s", err)
	}
	fixturePath = path.Join(wd, "../event/test-fixtures")
	files, err = ioutil.ReadDir(fixturePath)
	for _, f := range files {
		cmd = exec.Command("cp", path.Join(fixturePath, f.Name()), gtmPath)
		_, err = cmd.Output()
		if err != nil {
			t.Fatalf("Unable to copy %s directory to %s", fixturePath, gtmPath)
		}
	}

	return rootPath, gtmPath, func() {
		env.Paths = savePaths
		if err = os.RemoveAll(rootPath); err != nil {
			fmt.Printf("Error removing %s dir, %s", rootPath, err)
		}
	}
}
