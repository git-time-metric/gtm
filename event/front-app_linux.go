package epoch

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"

	ps "github.com/mitchellh/go-ps"
)

var (
	cmdPath string
	once    sync.Once
)

func getCommandPath() (string, error) {
	var err error
	once.Do(
		func() {
			cmdPath, err = exec.LookPath("xdotool")
		})
	return cmdPath, err
}

func getFrontApp() (string, error) {
	c, err := getCommandPath()
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
