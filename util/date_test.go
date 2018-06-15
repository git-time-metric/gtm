// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/now"
)

func printDates() {
	fmt.Printf("%+10s %s\n", "Today", TodayRange())
	fmt.Printf("%+10s %s\n", "Yesterday", YesterdayRange())
	fmt.Printf("%+10s %s\n", "ThisWeek", ThisWeekRange())
	fmt.Printf("%+10s %s\n", "LastWeek", LastWeekRange())
	fmt.Printf("%+10s %s\n", "ThisMonth", ThisMonthRange())
	fmt.Printf("%+10s %s\n", "LastMonth", LastMonthRange())
	fmt.Printf("%+10s %s\n", "ThisYear", ThisYearRange())
	fmt.Printf("%+10s %s\n", "LastYear", LastYearRange())
}

func TestDateRanges(t *testing.T) {
	tm, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}
	saveNow := Now
	defer func() { Now = saveNow }()
	Now = func() time.Time { return tm }

	TodayStart := "Wed Jul  1 00:00:00 UTC 2015"
	TodayEnd := "Wed Jul  1 23:59:59.999999999 UTC 2015"
	YesterdayStart := "Tue Jun 30 00:00:00 UTC 2015"
	YesterdayEnd := "Tue Jun 30 23:59:59.999999999 UTC 2015"
	ThisWeekStart := "Sun Jun 28 00:00:00 UTC 2015"
	ThisWeekEnd := "Sat Jul  4 23:59:59.999999999 UTC 2015"
	LastWeekStart := "Sun Jun 21 00:00:00 UTC 2015"
	LastWeekEnd := "Sat Jun 27 23:59:59.999999999 UTC 2015"
	ThisMonthStart := "Wed Jul  1 00:00:00 UTC 2015"
	ThisMonthEnd := "Fri Jul 31 23:59:59.999999999 UTC 2015"
	LastMonthStart := "Mon Jun  1 00:00:00 UTC 2015"
	LastMonthEnd := "Tue Jun 30 23:59:59.999999999 UTC 2015"
	ThisYearStart := "Thu Jan  1 00:00:00 UTC 2015"
	ThisYearEnd := "Thu Dec 31 23:59:59.999999999 UTC 2015"
	LastYearStart := "Wed Jan  1 00:00:00 UTC 2014"
	LastYearEnd := "Wed Dec 31 23:59:59.999999999 UTC 2014"

	dr := TodayRange()
	if !dr.Start.Equal(parseUnixDate(TodayStart, t)) || !dr.End.Equal(parseUnixDate(TodayEnd, t)) {
		t.Errorf("Today -> want %s - %s, got %s - %s", TodayStart, TodayEnd, dr.Start, dr.End)
	}

	dr = YesterdayRange()
	if !dr.Start.Equal(parseUnixDate(YesterdayStart, t)) || !dr.End.Equal(parseUnixDate(YesterdayEnd, t)) {
		t.Errorf("Yesterday -> want %s - %s, got %s - %s", YesterdayStart, YesterdayEnd, dr.Start, dr.End)
	}

	dr = ThisWeekRange()
	if !dr.Start.Equal(parseUnixDate(ThisWeekStart, t)) || !dr.End.Equal(parseUnixDate(ThisWeekEnd, t)) {
		t.Errorf("ThisWeek -> want %s - %s, got %s - %s", ThisWeekStart, ThisWeekEnd, dr.Start, dr.End)
	}

	dr = LastWeekRange()
	if !dr.Start.Equal(parseUnixDate(LastWeekStart, t)) || !dr.End.Equal(parseUnixDate(LastWeekEnd, t)) {
		t.Errorf("LastWeek -> want %s - %s, got %s - %s", LastWeekStart, LastWeekEnd, dr.Start, dr.End)
	}

	dr = ThisMonthRange()
	if !dr.Start.Equal(parseUnixDate(ThisMonthStart, t)) || !dr.End.Equal(parseUnixDate(ThisMonthEnd, t)) {
		t.Errorf("ThisMonth -> want %s - %s, got %s - %s", ThisMonthStart, LastWeekEnd, dr.Start, dr.End)
	}

	dr = LastMonthRange()
	if !dr.Start.Equal(parseUnixDate(LastMonthStart, t)) || !dr.End.Equal(parseUnixDate(LastMonthEnd, t)) {
		t.Errorf("LastMonth -> want %s - %s, got %s - %s", LastMonthStart, LastWeekEnd, dr.Start, dr.End)
	}

	dr = ThisYearRange()
	if !dr.Start.Equal(parseUnixDate(ThisYearStart, t)) || !dr.End.Equal(parseUnixDate(ThisYearEnd, t)) {
		t.Errorf("ThisYear -> want %s - %s, got %s - %s", ThisYearStart, LastWeekEnd, dr.Start, dr.End)
	}

	dr = LastYearRange()
	if !dr.Start.Equal(parseUnixDate(LastYearStart, t)) || !dr.End.Equal(parseUnixDate(LastYearEnd, t)) {
		t.Errorf("LastYear -> want %s - %s, got %s - %s", LastYearStart, LastWeekEnd, dr.Start, dr.End)
	}

}

func parseUnixDate(dt string, t *testing.T) time.Time {
	tm, err := time.Parse(time.UnixDate, dt)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestTodayRange(t *testing.T) {
	validDates := []time.Time{
		now.BeginningOfDay(),
		now.EndOfDay(),
		now.BeginningOfDay().Add(time.Nanosecond),
		now.EndOfDay().Add(-time.Nanosecond)}

	dateRange := TodayRange()

	for n, d := range validDates {
		if !dateRange.Within(d) {
			t.Errorf("%d: %s not within date range %+v", n, d, dateRange)
		}
	}

	invalidDates := []time.Time{
		now.BeginningOfDay().Add(-time.Nanosecond),
		now.EndOfDay().Add(time.Nanosecond),
		YesterdayRange().Start,
		YesterdayRange().End}

	for n, d := range invalidDates {
		if dateRange.Within(d) {
			t.Errorf("%d: %s is within date range %+v", n, d, dateRange)
		}
	}
}

func TestAfterNow(t *testing.T) {
	tm, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}
	saveNow := Now
	defer func() { Now = saveNow }()
	Now = func() time.Time { return tm }

	dt, err := time.Parse("2006-Jan-02", "2015-Jun-30")
	if err != nil {
		t.Fatal(err)
	}

	dr := AfterNow(0)
	if !dr.Within(dt) {
		t.Errorf("AfterNow(0) %s is not within date range %+v", dt, dr)
	}

	dr = AfterNow(1)
	if !dr.Within(dt) {
		t.Errorf("AfterNow(1) %s is not within date range %+v", dt, dr)
	}

	dr = AfterNow(2)
	if dr.Within(dt) {
		t.Errorf("AfterNow(2) %s is within date range %+v", dt, dr)
	}

}

func TestStartOnlyRange(t *testing.T) {
	tm, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}
	dr := DateRange{Start: tm}

	testDate, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}

	if !dr.Within(testDate) {
		t.Errorf("dr.Within(%s) not within %+v", testDate, dr)
	}

	testDate, err = time.Parse("2006-Jan-02", "2015-Jul-02")
	if err != nil {
		t.Fatal(err)
	}
	if !dr.Within(testDate) {
		t.Errorf("dr.Within(%s) not within %+v", testDate, dr)
	}

	testDate, err = time.Parse("2006-Jan-02", "2015-Jun-02")
	if err != nil {
		t.Fatal(err)
	}
	if dr.Within(testDate) {
		t.Errorf("dr.Within(%s) within %+v", testDate, dr)
	}
}

func TestEndOnlyRange(t *testing.T) {
	tm, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}
	dr := DateRange{End: tm}

	testDate, err := time.Parse("2006-Jan-02", "2015-Jul-01")
	if err != nil {
		t.Fatal(err)
	}

	if !dr.Within(testDate) {
		t.Errorf("dr.Within(%s) not within %+v", testDate, dr)
	}

	testDate, err = time.Parse("2006-Jan-02", "2015-Jun-30")
	if err != nil {
		t.Fatal(err)
	}
	if !dr.Within(testDate) {
		t.Errorf("dr.Within(%s) not within %+v", testDate, dr)
	}

	testDate, err = time.Parse("2006-Jan-02", "2015-Jul-02")
	if err != nil {
		t.Fatal(err)
	}
	if dr.Within(testDate) {
		t.Errorf("dr.Within(%s) within %+v", testDate, dr)
	}
}
