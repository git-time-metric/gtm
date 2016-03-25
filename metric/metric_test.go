package metric

import (
	"reflect"
	"testing"
)

var testdata string = `
1458681900:map[vim.log:1] 
1458682260:map[cmd/commit.go:1] 
1458683340:map[event/event.go:1] 
1458682740:map[vim.log:1] 
1458676380:map[scm/git.go:1] 
1458682440:map[event/event.go:2 metric/metric.go:1] 
1458683460:map[vim.log:2] 
1458682200:map[event/event.go:2 cmd/commit.go:2] 
1458911580:map[metric/metric.go:1] 
1458683400:map[event/event.go:4 event/event_test.go:2] 
1458911460:map[metric/metric.go:2] 
1458682320:map[cmd/commit.go:1] 
1458682500:map[event/event.go:1] 
1458682560:map[event/event.go:2] 
1458681780:map[vim.log:1] 
1458682080:map[event/event.go:1] 
1458682620:map[vim.log:1] 
1458683580:map[vim.log:1] 
1458676500:map[scm/git.go:1] 
1458676560:map[scm/git.go:1] 
1458682020:map[cmd/record.go:1 env/env.go:1] 
1458682140:map[cmd/record.go:2 event/event.go:2] 
1458911520:map[metric/metric.go:1] 
1458676440:map[scm/git.go:1] 
1458681840:map[vim.log:1] 
1458682680:map[vim.log:1] 
1458911400:map[event/event_test.go:2] 
1458683520:map[vim.log:1]]

6f53bc90ba625b5afaac80b422b44f1f609d6367:{Updated:true GitFile:event/event.go Time:380} 
fd3de0b7135021cc4c5ef23b8bea9ff98b704c47:{Updated:true GitFile:scm/git.go Time:240} 
26c5bdda12d74ceb9cf191911a79454bccd80640:{Updated:true GitFile:metric/metric.go Time:200} 
e65b42b6bf1eda6349451b063d46134dd7ab9921:{Updated:true GitFile:event/event_test.go Time:80} 
f93cea510c5049ff60ef12c62825a53f7d6e7d48:{Updated:true GitFile:cmd/record.go Time:60} 
1301df137d0acac0abf8cdc29bb74ef39ad2b042:{Updated:true GitFile:env/env.go Time:30} 
2dbf769f7faf2f921b89f3ff9d81d7b5e02a17a5:{Updated:true GitFile:vim.log Time:540} 
c2369545266e4a15c3db04a9f52b021364330bb7:{Updated:true GitFile:cmd/commit.go Time:150}]
`

func TestAllocateTime(t *testing.T) {
	cases := []struct {
		metric   map[string]metricFile
		event    map[string]int
		expected map[string]metricFile
	}{
		{
			map[string]metricFile{},
			map[string]int{"event/event.go": 1},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 60}},
		},
		{
			map[string]metricFile{},
			map[string]int{"event/event.go": 4, "event/event_test.go": 2},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 40},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 20}},
		},
		{
			map[string]metricFile{"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 60}},
			map[string]int{"event/event.go": 4, "event/event_test.go": 2},
			map[string]metricFile{
				"6f53bc90ba625b5afaac80b422b44f1f609d6367": metricFile{Updated: true, GitFile: "event/event.go", Time: 40},
				"e65b42b6bf1eda6349451b063d46134dd7ab9921": metricFile{Updated: true, GitFile: "event/event_test.go", Time: 80}},
		},
	}

	for _, tc := range cases {
		// copy metric map because it's updated in place during testing
		metricOrig := map[string]metricFile{}
		for k, v := range tc.metric {
			metricOrig[k] = v

		}
		allocateTime(tc.metric, tc.event)
		if !reflect.DeepEqual(tc.metric, tc.expected) {
			t.Errorf("allocateTime(%+v, %+v)\n want %+v\n got  %+v\n", metricOrig, tc.event, tc.expected, tc.metric)
		}
	}
}
