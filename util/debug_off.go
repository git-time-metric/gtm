// +build !debug

// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"io/ioutil"
	"log"
)

// Debug is no-op implementation of the debug logger
var Debug = log.New(ioutil.Discard, "", log.LstdFlags)
