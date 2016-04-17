package epoch

import "edgeg.io/gtm/project"

const WindowSize = 60

var IdleTimeout int64 = 120

func Minute(t int64) int64 {
	return (t / int64(WindowSize)) * WindowSize
}

func MinuteNow() int64 {
	return (project.Now().Unix() / int64(WindowSize)) * WindowSize
}

func Now() int64 {
	return project.Now().Unix()
}
