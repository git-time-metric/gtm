// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package epoch

import (
	"testing"
	"time"

	"github.com/git-time-metric/gtm/util"
)

func TestMinute(t *testing.T) {
	m := Minute(59)
	if m != 0 {
		t.Errorf("want 0 got %d", m)
	}
	m = Minute(61)
	if m != 60 {
		t.Errorf("want 60 got %d", m)
	}
	m = Minute(119)
	if m != 60 {
		t.Errorf("want 60 got %d", m)
	}
	m = Minute(120)
	if m != 120 {
		t.Errorf("want 120 got %d", m)
	}
}

func TestMinuteNow(t *testing.T) {
	tm, err := time.Parse("2006-01-02T15:04:05.999999999", "1970-01-01T00:04:05.999999999")
	if err != nil {
		t.Fatal(err)
	}
	saveNow := util.Now
	defer func() { util.Now = saveNow }()
	util.Now = func() time.Time { return tm }
	m := MinuteNow()
	if m != 240 {
		t.Errorf("want 240 got %d", m)
	}
}
