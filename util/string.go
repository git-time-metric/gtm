package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func Percent(val, total int) float64 {
	if total == 0 {
		return float64(0)
	}
	return (float64(val) / float64(total)) * 100
}

// FormatDuration converts seconds into a duration string
func FormatDuration(secs int) string {
	vals := regexp.MustCompile(`\d+`)
	matches := vals.FindAllString(DurationStr(secs), -1)
	switch len(matches) {
	case 3:
		return fmt.Sprintf("%sh %2sm %2ss", matches[0], matches[1], matches[2])
	case 2:
		return fmt.Sprintf("%sm %2ss", matches[0], matches[1])
	case 1:
		return fmt.Sprintf("%ss", matches[0])
	default:
		return ""
	}
}

//DurationStr returns seconds as a duration string, i.e. 9h10m30s
func DurationStr(secs int) string {
	return (time.Duration(secs) * time.Second).String()
}

// https://github.com/DaddyOh/golang-samples/blob/master/pad.go

// RightPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// LeftPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

// StringInSlice https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func StringInSlice(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
