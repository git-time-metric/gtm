package project

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"edgeg.io/gtm/scm"
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

// Now is the func used for system time within gtm
// This allows for manipulating system time during testing
var Now = func() time.Time { return time.Now() }

// Initialize initializes a git repo for time tracking
func Initialize() (string, error) {
	var fp string

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fp = filepath.Join(wd, ".git")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", wd)
	}

	fp = filepath.Join(wd, GTMDir)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0700); err != nil {
			return "", err
		}
	}

	if err := scm.SetHooks(GitHooks); err != nil {
		return "", err
	}

	if err := scm.Config(GitConfig); err != nil {
		return "", err
	}

	if err := scm.Ignore(GitIgnore); err != nil {
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
			wd,
			GitHooks,
			GitConfig,
			GitIgnore})

	if err != nil {
		return "", err
	}

	return b.String(), nil
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
