package event

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	cmdPath string
	once    sync.Once
)

func getCommandPath() (string, error) {
	var err error
	once.Do(
		func() {
			cmdPath, err = exec.LookPath("osascript")
		})
	return cmdPath, err
}

func getFrontApp() (string, error) {
	c, err := getCommandPath()
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
}
