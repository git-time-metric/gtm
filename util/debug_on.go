// +build debug

// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

// Debug is a debug logger
var Debug = NewContextLogger(log.New(ioutil.Discard, "debug ", log.Lmicroseconds), 3)

func init() {
	u, err := user.Current()
	if err != nil {
		return
	}

	w, err := os.OpenFile(filepath.Join(u.HomeDir, "gtm-debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	Debug = NewContextLogger(log.New(w, "debug ", log.Lmicroseconds), 3)
}
