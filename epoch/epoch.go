// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package epoch

import "github.com/git-time-metric/gtm/util"

// WindowSize is number seconds in an epoch window
const WindowSize = 60

// IdleTimeout seconds to record idle editor events
var IdleTimeout int64 = 120

// IdleProjectTimeout seconds to record application events without any editor events
var IdleProjectTimeout int64 = 300

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
