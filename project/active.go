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

var (
	logFile = "active-project.txt"
)

func SetActive(path string) error {
	logPath, err := activeProjectLog()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		logPath,
		[]byte(fmt.Sprintf("%s,%d", path, epoch.Now())), 0644); err != nil {
		return err
	}
	return nil
}

func GetActive() string {
	logPath, err := activeProjectLog()
	if err != nil {
		return ""
	}

	b, err := ioutil.ReadFile(logPath)
	if err != nil {
		return ""
	}

	parts := strings.Split(string(b), ",")
	if len(parts) != 2 {
		return ""
	}

	// does the project path exist
	if _, err := os.Stat(parts[0]); os.IsNotExist(err) {
		return ""
	}

	if !isActive(parts[1]) {
		return ""
	}

	return parts[0]
}

func isActive(timeUpdated string) bool {
	x, err := strconv.ParseInt(timeUpdated, 10, 64)
	if err != nil {
		return false
	}

	return time.Unix(epoch.Now(), 0).Before(time.Unix(x+epoch.IdleProjectTimeout, 0))
}

func activeProjectLog() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Join(u.HomeDir, gtmHomeDir), logFile), nil
}
