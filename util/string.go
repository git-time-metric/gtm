package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func FormatDuration(secs int) string {
	durStr := (time.Duration(secs) * time.Second).String()
	vals := regexp.MustCompile(`\d+`)
	matches := vals.FindAllString(durStr, -1)
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

// https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}
