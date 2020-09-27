// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package epoch

import "github.com/kilpkonn/gtm-enhanced/util"

// WindowSize is number seconds in an epoch window
const WindowSize = 60

// IdleTimeout is the number of seconds to record idle events for
var IdleTimeout int64 = 120

// Minute rounds epoch seconds down to the nearst epoch minute
func Minute(t int64) int64 {
	return (t / int64(WindowSize)) * WindowSize
}

// MinuteNow returns the epoch minute for the current time
func MinuteNow() int64 {
	return Minute(util.Now().Unix())
}

// Now returns the current Unix time
func Now() int64 {
	return util.Now().Unix()
}
