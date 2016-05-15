package scm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GitRootPath(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}

	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Unable to parse repository path, %s %s", string(b), err)
	}

	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("Unable to parse repository path, %s", err)
	}

	return s, nil
}

func GitBranch(wd ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to parse branch name, %s %s", string(b), err)
	}
	return strings.TrimSpace(string(b)), nil

}

func GitEmail(wd ...string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "user.email")
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to get user email, %s %s", string(b), err)
	}
	return strings.TrimSpace(string(b)), nil
}

func GitLastLog(wd ...string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--oneline", "--raw")
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		// if there are no git commits yet it will fail
		// ignoring this error
		return "", nil
	}
	return string(b), err
}

func GitLogSHA1s(args []string, wd ...string) ([]string, error) {
	args = append([]string{"log", "--pretty=%H"}, args...)
	cmd := exec.Command("git", args...)
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return []string{}, fmt.Errorf("Unable to get SHA1s, %s %s", string(b), err)
	}
	parts := strings.Split(string(b), "\n")
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
	}
	return parts, err
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

func GitAddNote(n string, nameSpace string, wd ...string) error {
	cmd := exec.Command("git", "notes", fmt.Sprintf("--ref=%s", nameSpace), "add", "-f", "-m", n)
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	if b, err := cmd.Output(); err != nil {
		return fmt.Errorf("Unable to add git note, %s %s", string(b), err)
	}
	return nil
}

func GitNote(commitID string, nameSpace string, wd ...string) (string, error) {
	cmd := exec.Command("git", "notes", fmt.Sprintf("--ref=%s", nameSpace), "show", commitID)
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Git note not found %s", string(b))
	}
	return string(b), nil
}

func GitLog(commitID string, wd ...string) ([]string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty", "--format=%an,%ad,%h,%s", commitID)
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return []string{}, fmt.Errorf("Unable to execute git log, %s %s", string(b), err)
	}
	parts := strings.SplitN(string(b), ",", 4)
	if len(parts) < 4 {
		return []string{}, fmt.Errorf("Unable to parse log output, %s %s", string(b), err)
	}
	return parts, err
}

func GitConfig(settings map[string]string, wd ...string) error {
	for k, v := range settings {
		cmd := exec.Command("git", "config", "-l")
		if len(wd) > 0 {
			cmd.Dir = wd[0]
		}
		var (
			b   []byte
			err error
		)
		if b, err = cmd.Output(); err != nil {
			return fmt.Errorf("Unable to run git config -l, %s %s", string(b), err)
		}
		if !strings.Contains(string(b), fmt.Sprintf("%s=%s", k, v)) {
			cmd := exec.Command("git", "config", "--add", k, v)
			if len(wd) > 0 {
				cmd.Dir = wd[0]
			}
			if b, err := cmd.Output(); err != nil {
				return fmt.Errorf("Unable to run git config --add %s %s, %s %s", k, v, string(b), err)
			}
		}
	}
	return nil
}

func GitTracked(f string, wd ...string) (bool, error) {
	cmd := exec.Command("git", "ls-files", f)
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git tracked status for %s, %s %s", f, string(b), err)
	}
	return strings.TrimSpace(string(b)) != "", nil
}

func GitModified(f string, staging bool, wd ...string) (bool, error) {
	var cmd *exec.Cmd
	if staging {
		cmd = exec.Command("git", "diff", "--name-only", "--cached", "--", f)
	} else {
		cmd = exec.Command("git", "diff", "--name-only", "--", f)
	}
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git modified status for %s, %s %s", f, string(b), err)
	}
	return strings.TrimSpace(string(b)) != "", nil
}

func GitHasStaged(wd ...string) (bool, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("git", "diff", "--name-only", "--cached")
	if len(wd) > 0 {
		cmd.Dir = wd[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git staged status, %s %s", string(b), err)
	}
	return strings.TrimSpace(string(b)) != "", nil
}

func GitSetHooks(hooks map[string]string, wd ...string) error {
	for hook, command := range hooks {
		var (
			p   string
			err error
		)

		if len(wd) > 0 {
			p = wd[0]
		} else {
			p, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		fp := path.Join(p, ".git", "hooks", hook)

		var output string
		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			b, err := ioutil.ReadFile(fp)
			if err != nil {
				return err
			}
			output = string(b)

			if strings.Contains(output, command+"\n") {
				// if file already exists this will make sure it's executable
				if err := os.Chmod(fp, 0755); err != nil {
					return err
				}
				return nil
			}
		}

		if err = ioutil.WriteFile(
			fp, []byte(fmt.Sprintf("%s\n%s\n", output, command)), 0755); err != nil {
			return err
		}
		// if file already exists this will make sure it's executable
		if err := os.Chmod(fp, 0755); err != nil {
			return err
		}
	}

	return nil
}

func GitIgnore(ignore string, wd ...string) error {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p = wd[0]
	} else {
		p, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	fp := path.Join(p, ".gitignore")

	var output string
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output = string(b)

		if strings.Contains(output, ignore+"\n") {
			return nil
		}
	}

	if err = ioutil.WriteFile(
		fp, []byte(fmt.Sprintf("%s\n%s\n", output, ignore)), 0644); err != nil {
		return err
	}
	return nil
}
