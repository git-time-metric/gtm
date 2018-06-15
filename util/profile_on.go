// +build profile

// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var Profile func(s ...string) func()

func init() {
	u, err := user.Current()
	if err != nil {
		return
	}

	w, err := os.OpenFile(filepath.Join(u.HomeDir, "gtm-profile.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	profileLog := NewContextLogger(log.New(w, "profile ", log.Lmicroseconds), 4)

	Profile = func(s ...string) func() {
		start := time.Now()
		label := ""
		if len(s) > 0 {
			label = s[0]
		}
		return func() {
			t := time.Since(start)
			if label == "" {
				profileLog.Printf("%s", t)
			} else {
				profileLog.Printf("%s %s", label, t)
			}
		}
	}
}
