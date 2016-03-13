package epoch

import "edgeg.io/gtm/env"

const (
	IdleTimeout = 120
	WindowSize  = 60
)

func MinuteNow() int64 {
	return (env.Now().Unix() / int64(WindowSize)) * WindowSize
}

func MinutePast() int64 {
	// go back a minute plus 5 more seconds
	// this prevents the potential of missing events recorded on the minute boudaries
	return ((env.Now().Unix() - 65) / int64(WindowSize)) * WindowSize
}
