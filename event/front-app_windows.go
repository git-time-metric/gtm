package event

import (
	"github.com/git-time-metric/w32"
	ps "github.com/mitchellh/go-ps"
)

func getFrontApp() (string, error) {
	hwnd := w32.GetWindowForeground()
	_, pid := w32.GetWindowThreadProcessId(hwnd)
	p, err := ps.FindProcess(pid)
	if err != nil {
		return "", err
	}

	return p.Executable(), nil
}
