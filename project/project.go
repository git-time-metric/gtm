package project

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/git-time-metric/gtm/scm"
)

var (
	// ErrNotInitialized is raised when a git repo not initialized for time tracking
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	// ErrFileNotFound is raised when record an event for a file that does not exist
	ErrFileNotFound = errors.New("File does not exist")
)

var (
	// GitHooks is map of hooks to apply to the git repo
	GitHooks = map[string]string{
		"post-commit": "gtm commit --yes"}
	// GitConfig is map of git configuration settings
	GitConfig = map[string]string{
		"alias.pushgtm":    "push origin refs/notes/gtm-data",
		"alias.fetchgtm":   "fetch origin refs/notes/gtm-data:refs/notes/gtm-data",
		"notes.rewriteref": "refs/notes/gtm-data"}
	// GitIgnore is list of file/path ignores to apply to git repo
	GitIgnore = ".gtm/"
)

const (
	// NoteNameSpace is the gtm git note namespace
	NoteNameSpace = "gtm-data"
	// GTMDir is the subdir for gtm within the git repo root directory
	GTMDir = ".gtm"
)

const initMsgTpl string = `
{{print "Git Time Metric initialized for " (.ProjectPath) | printf (.HeaderFormat) }}

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%16s" }}: {{ $command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%16s" }}: {{ $val }}
{{end -}}
{{ print ".gitignore:" | printf "%17s" }} {{ .GitIgnore }}
`
const removeMsgTpl string = `
{{print "Git Time Metric uninitialized for " (.ProjectPath) | printf (.HeaderFormat) }}

The following items have been removed.

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%16s" }}: {{ $command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%16s" }}: {{ $val }}
{{end -}}
{{ print ".gitignore:" | printf "%17s" }} {{ .GitIgnore }}
`

// Now is the func used for system time within gtm
// This allows for manipulating system time during testing
var Now = func() time.Time { return time.Now() }

// Initialize initializes a git repo for time tracking
func Initialize() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", projRoot)
	}

	gitPath := filepath.Join(projRoot, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", gitPath)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		if err := os.MkdirAll(gtmPath, 0700); err != nil {
			return "", err
		}
	}

	if err := scm.SetHooks(GitHooks, projRoot); err != nil {
		return "", err
	}

	if err := scm.ConfigSet(GitConfig, projRoot); err != nil {
		return "", err
	}

	if err := scm.IgnoreSet(GitIgnore, projRoot); err != nil {
		return "", err
	}

	headerFormat := "%s"
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(initMsgTpl))
	err = t.Execute(b,
		struct {
			HeaderFormat string
			ProjectPath  string
			GitHooks     map[string]string
			GitConfig    map[string]string
			GitIgnore    string
		}{
			headerFormat,
			projRoot,
			GitHooks,
			GitConfig,
			GitIgnore})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//Uninitialize remove GTM tracking from the project in the current working directory
func Uninitialize() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to unintialize Git Time Metric, Git repository not found in %s", projRoot)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to uninitialize Git Time Metric, %s directory not found", gtmPath)
	}
	if err := scm.RemoveHooks(GitHooks, projRoot); err != nil {
		return "", err
	}
	if err := scm.ConfigRemove(GitConfig, projRoot); err != nil {
		return "", err
	}
	if err := scm.IgnoreRemove(GitIgnore, projRoot); err != nil {
		return "", err
	}
	if err := os.RemoveAll(gtmPath); err != nil {
		return "", err
	}

	headerFormat := "%s"
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(removeMsgTpl))
	err = t.Execute(b,
		struct {
			HeaderFormat string
			ProjectPath  string
			GitHooks     map[string]string
			GitConfig    map[string]string
			GitIgnore    string
		}{
			headerFormat,
			projRoot,
			GitHooks,
			GitConfig,
			GitIgnore})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//Clean removes any event or metrics files from project in the current working directory
func Clean() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to clean, Git repository not found in %s", projRoot)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to clean GTM data, %s directory not found", gtmPath)
	}

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".event") &&
			!strings.HasSuffix(f.Name(), ".metric") {
			continue
		}
		if err := os.Remove(filepath.Join(gtmPath, f.Name())); err != nil {
			return "", err
		}
	}

	return "", nil
}

// Paths returns the root git repo and gtm paths
func Paths(wd ...string) (string, string, error) {
	var (
		repoPath string
		err      error
	)
	if len(wd) > 0 {
		repoPath, err = scm.RootPath(wd[0])
	} else {
		repoPath, err = scm.RootPath()
	}
	if err != nil {
		return "", "", ErrNotInitialized
	}

	gtmPath := filepath.Join(repoPath, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", "", ErrNotInitialized
	}
	return repoPath, gtmPath, nil
}

// Log logs to a gtm log in the GTMDir
func Log(v ...interface{}) error {
	_, gtmPath, err := Paths()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(gtmPath, "gtm.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer func() { _ = f.Close() }()
	log.SetOutput(f)

	log.Println(v)
	return nil
}
