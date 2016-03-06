package cfg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	GitMetricDir = ".gmetric"
)

var (
	ErrGitMetricNotInitialized = errors.New("Git Metric is not initialized")
)

func Initialize() error {
	var fp string

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fp = filepath.Join(wd, ".git")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return fmt.Errorf(
			"Unable to intialize GitMetric, Git repository not found in %s", wd)
	}

	fp = filepath.Join(wd, GitMetricDir)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0700); err != nil {
			return err
		}
	}

	return nil
}
