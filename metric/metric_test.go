// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package metric

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestAllocateTime(t *testing.T) {
	cases := []struct {
		metric   map[string]FileMetric
		event    map[string]int
		expected map[string]FileMetric
	}{
		{
			map[string]FileMetric{},
			map[string]int{filepath.Join("event", "event.go"): 1},
			map[string]FileMetric{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": {Updated: true, SourceFile: filepath.Join("event", "event.go"), TimeSpent: 60, Timeline: map[int64]int{int64(1): 60}}},
		},
		{
			map[string]FileMetric{},
			map[string]int{filepath.Join("event", "event.go"): 4, filepath.Join("event", "event_test.go"): 2},
			map[string]FileMetric{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": {Updated: true, SourceFile: filepath.Join("event", "event.go"), TimeSpent: 40, Timeline: map[int64]int{int64(1): 40}},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": {Updated: true, SourceFile: filepath.Join("event", "event_test.go"), TimeSpent: 20, Timeline: map[int64]int{int64(1): 20}}},
		},
		{
			map[string]FileMetric{"e65b42b6bf1eda6349451b063d46134dd7ab9921": {Updated: true, SourceFile: filepath.Join("event", "event_test.go"), TimeSpent: 60, Timeline: map[int64]int{int64(1): 60}}},
			map[string]int{filepath.Join("event", "event.go"): 4, filepath.Join("event", "event_test.go"): 2},
			map[string]FileMetric{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": {Updated: true, SourceFile: filepath.Join("event", "event.go"), TimeSpent: 40, Timeline: map[int64]int{int64(1): 40}},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": {Updated: true, SourceFile: filepath.Join("event", "event_test.go"), TimeSpent: 80, Timeline: map[int64]int{int64(1): 80}}},
		},
	}

	for _, tc := range cases {
		// copy metric map because it's updated in place during testing
		metricOrig := map[string]FileMetric{}
		for k, v := range tc.metric {
			metricOrig[k] = v

		}
		if err := allocateTime(1, tc.metric, tc.event); err != nil {
			t.Errorf("allocateTime(%+v, %+v) want error nil got %s", metricOrig, tc.event, err)
		}

		if !reflect.DeepEqual(tc.metric, tc.expected) {
			t.Errorf("allocateTime(%+v, %+v)\nwant:\n%+v\ngot:\n%+v\n", metricOrig, tc.event, tc.expected, tc.metric)
		}
	}
}

func TestFileID(t *testing.T) {
	want := "6f53bc90ba625b5afaac80b422b44f1f609d6367"
	got := getFileID(filepath.Join("event", "event.go"))
	if want != got {
		t.Errorf("getFileID(%s), want %s, got %s", filepath.Join("event", "event.go"), want, got)

	}
}
