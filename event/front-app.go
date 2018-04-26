package event

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	ps "github.com/mitchellh/go-ps"
)

var (
	cmdPath string
	once    sync.Once
)

func getCommandPath(x string) (string, error) {
	var err error
	once.Do(
		func() {
			cmdPath, err = exec.LookPath(x)
		})
	return cmdPath, err
}

func getFrontApp() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		c, err := getCommandPath("osascript")
		if err != nil {
			return "", err
		}

		x := exec.Command(c,
			`-e`, `tell application "System Events"`,
			`-e`, `set frontApp to name of first application process whose frontmost is true`,
			`-e`, `end tell`,
		)

		o, err := x.CombinedOutput()
		if err != nil {
			return "", err
		}

		return normalizeAppName(strings.Replace(string(o), "\n", "", -1)), nil

	case "linux":
		c, err := getCommandPath("xdotool")
		if err != nil {
			return "", err
		}

		x := exec.Command(c, "getwindowfocus", "getwindowpid")

		o, err := x.CombinedOutput()
		if err != nil {
			return "", err
		}

		pid, err := strconv.Atoi(strings.Replace(string(o), "\n", "", -1))
		if err != nil {
			return "", err
		}

		p, err := ps.FindProcess(pid)
		if err != nil {
			return "", err
		}

		return p.Executable(), nil
	}
	return "", errors.New(fmt.Sprintf("%s not supported", runtime.GOOS))
}
