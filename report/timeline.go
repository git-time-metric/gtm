// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"sort"
	"strconv"
	"time"

	"github.com/git-time-metric/gtm/util"
)

type timelineCommitEntries []timelineCommitEntry

type timelineCommitEntry struct {
	Day     string
	Total   int
	Commits [24]int
}

func (t *timelineCommitEntry) inc(hour int) {
	t.Total++
	t.Commits[hour]++
}

func (t timelineCommitEntries) HourMaxCommits() int {
	max := 0
	for _, entry := range t {
		for _, commits := range entry.Commits {
			if commits > max {
				max = commits
			}
		}
	}
	return max
}

func (t timelineCommitEntries) Total() int {
	total := 0
	for _, e := range t {
		total += e.Total
	}
	return total
}

func (c commitNoteDetails) timelineCommits() (timelineCommitEntries, error) {
	timelineMap := map[string]timelineCommitEntry{}
	timeline := []timelineCommitEntry{}
	for _, n := range c {
		t := n.When
		day := t.Format("2006-01-02")
		hour, err := strconv.Atoi(t.Format("15"))
		if err != nil {
			return timelineCommitEntries{}, err
		}
		if entry, ok := timelineMap[day]; !ok {
			var commits [24]int
			commits[hour] = 1
			timelineMap[day] = timelineCommitEntry{Day: t.Format("Mon Jan 02"), Commits: commits, Total: 1}
		} else {
			entry.inc(hour)
			timelineMap[day] = entry
		}
	}

	keys := make([]string, 0, len(timelineMap))
	for key := range timelineMap {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, k := range keys {
		timeline = append(timeline, timelineMap[k])
	}

	return timeline, nil
}

func (c commitNoteDetails) timeline() (timelineEntries, error) {
	timelineMap := map[string]timelineEntry{}
	timeline := []timelineEntry{}
	for _, n := range c {
		for _, f := range n.Note.Files {
			for epoch, secs := range f.Timeline {
				t := time.Unix(epoch, 0)
				day := t.Format("2006-01-02")
				hour, err := strconv.Atoi(t.Format("15"))
				if err != nil {
					return timelineEntries{}, err
				}
				if entry, ok := timelineMap[day]; !ok {
					var hours [24]int
					hours[hour] = secs
					timelineMap[day] = timelineEntry{Day: t.Format("Mon Jan 02"), Hours: hours, Seconds: secs}
				} else {
					entry.add(secs, hour)
					timelineMap[day] = entry
				}
			}
		}
	}

	keys := make([]string, 0, len(timelineMap))
	for key := range timelineMap {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, k := range keys {
		timeline = append(timeline, timelineMap[k])
	}

	return timeline, nil
}

type timelineEntries []timelineEntry

func (t timelineEntries) Duration() string {
	total := 0
	for _, entry := range t {
		total += entry.Seconds
	}
	return util.FormatDuration(total)
}

func (t timelineEntries) HourMaxSeconds() int {
	// Default to number of seconds in a hour
	// Actual max can be much higher when reporting
	// across multiple projects and users
	max := 3600
	for _, entry := range t {
		for _, secs := range entry.Hours {
			if secs > max {
				max = secs
			}
		}
	}
	return max
}

type timelineEntry struct {
	Day     string
	Seconds int
	Hours   [24]int
}

func (t *timelineEntry) add(s int, hour int) {
	t.Seconds += s
	t.Hours[hour] += s
}

func (t *timelineEntry) Duration() string {
	return util.FormatDuration(t.Seconds)
}
