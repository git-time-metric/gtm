package env

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"edgeg.io/gtm/scm"
)

var (
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	ErrFileNotFound   = errors.New("File does not exist")
)

var (
	NoteNameSpace  string = "gtm-data"
	GTMDirectory   string = ".gtm"
	PostCommitHook string = "gtm commit --dry-run=false"
)

var Now = func() time.Time { return time.Now() }

func Initialize() error {
	var fp string

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fp = filepath.Join(wd, ".git")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", wd)
	}

	fp = filepath.Join(wd, GTMDirectory)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0700); err != nil {
			return err
		}
	}

	if err := scm.GitInitHook("post-commit", PostCommitHook); err != nil {
		return err
	}

	return nil
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
	defer f.Close()
	log.SetOutput(f)

	log.Println(v)
	return nil
}
