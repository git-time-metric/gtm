package epoch

import "time"

const (
	IdleTimeout = 120
	WindowSize  = 60
)

var (
	overrideNow bool = false
	timeNow     time.Time
)

func MinuteNow() int64 {
	return (Now().Unix() / int64(WindowSize)) * WindowSize
}

func MinutePast() int64 {
	// go back a minute plus 5 more seconds
	// this prevents the potential of missing events recorded on the minute boudaries
	return ((Now().Unix() - 65) / int64(WindowSize)) * WindowSize
}

func SetNow(t time.Time) {
	overrideNow = true
	timeNow = t
}

func ClearNow() {
	overrideNow = false
}

func Now() time.Time {
	if overrideNow {
		return timeNow
	}
	return time.Now()
}
