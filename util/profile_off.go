// +build !profile

// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

var (
	Profile = func(s ...string) func() {
		return func() {
		}
	}
)
