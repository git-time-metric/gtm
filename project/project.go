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
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	ErrFileNotFound   = errors.New("File does not exist")
)

var (
	NoteNameSpace = "gtm-data"
	GTMDirectory  = ".gtm"
	GitHooks      = map[string]string{
		"post-commit": "gtm commit --yes"}
	GitConfig = map[string]string{
		"alias.pushgtm":    "push origin refs/notes/gtm-data",
		"alias.fetchgtm":   "fetch origin refs/notes/gtm-data:refs/notes/gtm-data",
		"notes.rewriteref": "refs/notes/gtm-data"}
	GitIgnore = ".gtm/"
)

const InitMsgTpl string = `
{{print "Git Time Metric initialized for " (.ProjectPath) | printf (.HeaderFormat) }}

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%16s" }}: {{ $command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%16s" }}: {{ $val }}
{{end -}}
{{ print ".gitignore:" | printf "%17s" }} {{ .GitIgnore }}
`

var Now = func() time.Time { return time.Now() }

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

	fp = filepath.Join(wd, GTMDirectory)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0700); err != nil {
			return "", err
		}
	}

	if err := scm.GitSetHooks(GitHooks); err != nil {
		return "", err
	}

	if err := scm.GitConfig(GitConfig); err != nil {
		return "", err
	}

	if err := scm.GitIgnore(GitIgnore); err != nil {
		return "", err
	}

	headerFormat := "%s"
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(InitMsgTpl))
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

// The Paths function returns the git repository root path and the gtm path within the root.
// If the path is not blank, it's used as the current working directory when retrieving the root path.
//
// Note - the function is declared as a variable to allow for mocking during testing.
//
var Paths = func(path ...string) (string, string, error) {
	p := ""
	if len(path) > 0 {
		p = path[0]
	}
	rootPath, err := scm.GitRootPath(p)
	if err != nil {
		return "", "", ErrNotInitialized
	}

	gtmPath := filepath.Join(rootPath, GTMDirectory)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", "", ErrNotInitialized
	}
	return rootPath, gtmPath, nil
}

func LogToGTM(v ...interface{}) error {
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
