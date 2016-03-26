package scm

import (
	"fmt"
	"os/exec"
	"strings"
)

func GitRootPath(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}

	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Unable to parse repository path, %s", err)
	}

	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("Unable to parse repository path, %s", err)
	}

	return s, nil
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

func GitCommitMsg(path ...string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--oneline", "--raw")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if b, err := cmd.Output(); err != nil {
		return "", nil
	} else {
		return string(b), err
	}
}

func GitParseMessage(m string) (uuid, msg string, files []string) {
	l := strings.Split(m, "\n")
	files = make([]string, 0)
	for i, v := range l {
		if i == 0 {
			s := strings.SplitN(v, " ", 2)
			uuid = s[0]
			msg = s[1]
		} else {
			if strings.TrimSpace(v) != "" {
				s := strings.Split(v, "\t")
				files = append(files, s[1])
			}
		}
	}
	return
}

func GitAddNote(n string, nameSpace string, path ...string) error {
	cmd := exec.Command("git", "notes", fmt.Sprintf("--ref=%s", nameSpace), "add", "-f", "-m", n)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("Unable to add git note %s", err)
	}
	return nil
}

func GitTracked(f string, path ...string) (bool, error) {
	cmd := exec.Command("git", "ls-files", f)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if out, err := cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git tracked status for %s, %s", f, err)
	} else {
		return strings.TrimSpace(string(out)) != "", nil
	}
}

func GitModified(f string, path ...string) (bool, error) {
	cmd := exec.Command("git", "ls-files", "-m", f)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if out, err := cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git modified status for %s, %s", f, err)
	} else {
		return strings.TrimSpace(string(out)) != "", nil
	}
}
