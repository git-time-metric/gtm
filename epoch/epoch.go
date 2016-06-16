package epoch

import "edgeg.io/gtm/project"

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
	return Minute(project.Now().Unix())
}
