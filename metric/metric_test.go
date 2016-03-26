package metric

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"edgeg.io/gtm/env"
)

var testdata string = `
1458681900:map[vim.log:1] 
1458682260:map[cmd/commit.go:1] 
1458683340:map[event/event.go:1] 
1458682740:map[vim.log:1] 
1458676380:map[scm/git.go:1] 
1458682440:map[event/event.go:2 metric/metric.go:1] 
1458683460:map[vim.log:2] 
1458682200:map[event/event.go:2 cmd/commit.go:2] 
1458911580:map[metric/metric.go:1] 
1458683400:map[event/event.go:4 event/event_test.go:2] 
1458911460:map[metric/metric.go:2] 
1458682320:map[cmd/commit.go:1] 
1458682500:map[event/event.go:1] 
1458682560:map[event/event.go:2] 
1458681780:map[vim.log:1] 
1458682080:map[event/event.go:1] 
1458682620:map[vim.log:1] 
1458683580:map[vim.log:1] 
1458676500:map[scm/git.go:1] 
1458676560:map[scm/git.go:1] 
1458682020:map[cmd/record.go:1 env/env.go:1] 
1458682140:map[cmd/record.go:2 event/event.go:2] 
1458911520:map[metric/metric.go:1] 
1458676440:map[scm/git.go:1] 
1458681840:map[vim.log:1] 
1458682680:map[vim.log:1] 
1458911400:map[event/event_test.go:2] 
1458683520:map[vim.log:1]]

6f53bc90ba625b5afaac80b422b44f1f609d6367:{Updated:true GitFile:event/event.go Time:380} 
fd3de0b7135021cc4c5ef23b8bea9ff98b704c47:{Updated:true GitFile:scm/git.go Time:240} 
26c5bdda12d74ceb9cf191911a79454bccd80640:{Updated:true GitFile:metric/metric.go Time:200} 
e65b42b6bf1eda6349451b063d46134dd7ab9921:{Updated:true GitFile:event/event_test.go Time:80} 
f93cea510c5049ff60ef12c62825a53f7d6e7d48:{Updated:true GitFile:cmd/record.go Time:60} 
1301df137d0acac0abf8cdc29bb74ef39ad2b042:{Updated:true GitFile:env/env.go Time:30} 
2dbf769f7faf2f921b89f3ff9d81d7b5e02a17a5:{Updated:true GitFile:vim.log Time:540} 
c2369545266e4a15c3db04a9f52b021364330bb7:{Updated:true GitFile:cmd/commit.go Time:150}]
`

func TestAllocateTime(t *testing.T) {
	cases := []struct {
		metric   map[string]metricFile
		event    map[string]int
		expected map[string]metricFile
	}{
		{
			map[string]metricFile{},
			map[string]int{"event/event.go": 1},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 60}},
		},
		{
			map[string]metricFile{},
			map[string]int{"event/event.go": 4, "event/event_test.go": 2},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 40},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 20}},
		},
		{
			map[string]metricFile{"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 60}},
			map[string]int{"event/event.go": 4, "event/event_test.go": 2},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 40},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 80}},
		},
	}

	for _, tc := range cases {
		// copy metric map because it's updated in place during testing
		metricOrig := map[string]metricFile{}
		for k, v := range tc.metric {
			metricOrig[k] = v

		}
		allocateTime(tc.metric, tc.event)
		if !reflect.DeepEqual(tc.metric, tc.expected) {
			t.Errorf("allocateTime(%+v, %+v)\n want %+v\n got  %+v\n", metricOrig, tc.event, tc.expected, tc.metric)
		}
	}
}

func TestFileID(t *testing.T) {
	want := "6f53bc90ba625b5afaac80b422b44f1f609d6367"
	got := getFileID("event/event.go")
	if want != got {
		t.Errorf("getFileID(%s), want %s, got %s", "event/event.go", want, got)

	}
}

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
	defer func() { os.Chdir(savedCurDir) }()
	os.Chdir(rootPath)
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

	err = Process(false)
	if err != nil {
		t.Fatalf("Process(false), want error nil, got %s", err)
	}

	cmd = exec.Command("git", "notes", "--ref", "gtm", "show")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git notes, %s", string(b))
	}

	want := []string{"total: 300", "event.go: 280", "event_test.go: 20"}
	for _, s := range want {
		if !strings.Contains(string(b), s) {
			t.Errorf("Process(false) - test full commit, \nwant \n%s, \ngot \n%s", s, string(b))
		}

	}

	// Test Process with committing only one of the two git tracked files that have been modified

	// change back to saved current working directory and setup
	os.Chdir(savedCurDir)
	rootPath, gtmPath, f2 := processSetup(t)
	defer f2()

	// chandge working directory and initialize git repo
	os.Chdir(rootPath)
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

	err = Process(false)
	if err != nil {
		t.Fatalf("Process(false), want error nil, got %s", err)
	}

	cmd = exec.Command("git", "notes", "--ref", "gtm", "show")
	b, err = cmd.Output()
	if err != nil {
		t.Fatalf("Unable to run git notes, %s", string(b))
	}

	want = []string{"total: 20", "event_test.go: 20"}
	for _, s := range want {
		if !strings.Contains(string(b), s) {
			t.Errorf("Process(false), \nwant \n%s, \ngot \n%s", s, string(b))
		}

	}
	p := path.Join(gtmPath, "6f53bc90ba625b5afaac80b422b44f1f609d6367.metric")
	if !env.FileExists(p) {
		t.Errorf("Process(false) - test partial commit, want file %s exist, got file exists false", p)
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
	gtmPath = path.Join(rootPath, ".gtm")
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
