package event

import (
	"errors"

	"github.com/AllenDang/w32"
	ps "github.com/mitchellh/go-ps"
)

func getFrontApp() (string, error) {
	hwnd := w32.GetWindowForeground()
	if hwnd == nil {
		return "", errors.New("Unable to get window in the foreground")
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	p, err := ps.FindProcess(pid)
	if err != nil {
		return "", err
	}

	return p.Executable(), nil
}
