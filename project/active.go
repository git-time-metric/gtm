// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/epoch"
)

// SetActive records a project's path as active
// A package level func variable is used for ease of testing.
var SetActive = func(path string) error {
	a := active{path: path, lastUpdated: epoch.Now()}
	err := a.marshal()
	if err != nil {
		// FIXME: do not eat error
		return nil
	}
	return nil
}

// GetActive returns the current active project's path.
// If not project is active an empty string is returned.
// A package level func variable is used for ease of testing.
var GetActive = func() string {
	a := active{}
	err := a.unmarshal()
	if err != nil {
		// FIXME: do not eat error
		return ""
	}

	if !a.pathExists() {
		return ""
	}

	if !time.Unix(epoch.Now(), 0).Before(
		time.Unix(a.lastUpdated+epoch.IdleProjectTimeout, 0)) {
		return ""
	}
	return a.path
}

type active struct {
	path        string
	lastUpdated int64
}

// ActiveSerializationPath returns the path to serialize the active project to.
// A package level func variable is used for ease of testing.
var ActiveSerializationPath = func() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Join(u.HomeDir, gtmHomeDir), "active-project.txt"), nil
}

func (a *active) pathExists() bool {
	if _, err := os.Stat(a.path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (a *active) marshal() error {
	f, err := ActiveSerializationPath()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		f, []byte(fmt.Sprintf("%s,%d", a.path, a.lastUpdated)), 0644); err != nil {
		return err
	}
	return nil
}

func (a *active) unmarshal() error {
	f, err := ActiveSerializationPath()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	parts := strings.Split(string(b), ",")
	if len(parts) != 2 {
		return err
	}

	a.path = parts[0]

	a.lastUpdated, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	return nil
}
