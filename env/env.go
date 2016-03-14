package env

import (
	"errors"
	"fmt"
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
	overrideNow bool = false
	timeNow     time.Time
)

func Initialize() error {
	//TODO initialize post git commit hook
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

	fp = filepath.Join(wd, ".gtm")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0700); err != nil {
			return err
		}
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

	gtmPath := filepath.Join(rootPath, ".gtm")
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", "", ErrNotInitialized
	}
	return rootPath, gtmPath, nil
}

func FilePath(f string) (string, error) {
	p := filepath.Dir(f)
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("Unable to extract file path from %s, %s", f, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("Unable to extract file path from %s", f)
	}
	return p, nil
}

func FileExists(f string) bool {
	if fileInfo, err := os.Stat(f); os.IsNotExist(err) || fileInfo.IsDir() {
		return false
	}
	return true
}

func SetNow(t time.Time) {
	overrideNow = true
	timeNow = t
}

func ClearNow() {
	overrideNow = false
}

func Now() time.Time {
	if overrideNow {
		return timeNow
	}
	return time.Now()
}
