package cfg

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	ErrFileNotFound   = errors.New("File does not exist")
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

func Paths(path ...string) (string, string, error) {
	p := ""
	if len(path) > 0 {
		p = path[0]
	}
	rootPath, err := GitRootPath(p)
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

func RelativePath(filePath, dirPath string) (string, error) {
	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		return "", fmt.Errorf("Unable to return a relativePath, path is not a valid directory %s", dirPath)
	}
	return strings.Replace(filePath, dirPath, "", 1), nil
}

func GitRootPath(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if b, err := cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to parse repository path, %s", err)
	} else {
		return strings.TrimSpace(string(b)), nil
	}
}

func GitBranch(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if b, err := cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to parse branch name, %s", err)
	} else {
		return strings.TrimSpace(string(b)), nil
	}
}

func GitEmail(path ...string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "user.email")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if b, err := cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to get user email, %s", err)
	} else {
		return strings.TrimSpace(string(b)), nil
	}
}

func FileExist(f string) bool {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveFiles(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}
