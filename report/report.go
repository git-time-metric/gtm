// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/git-time-metric/gtm/note"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/util"
	isatty "github.com/mattn/go-isatty"
)

var funcMap = template.FuncMap{
	"FormatDuration": util.FormatDuration,
	"RightPad2Len":   util.RightPad2Len,
	"LeftPad2Len":    util.LeftPad2Len,
	"Percent":        util.Percent,
	"Blocks":         BlockForVal,
}

// ProjectCommits contains a project's directory path and commit ids
type ProjectCommits struct {
	Path    string
	Commits []string
}

// OutputOptions contains cli options for reporting
type OutputOptions struct {
	TotalOnly   bool
	FullMessage bool
	TerminalOff bool
	Color       bool
	Limit       int
}

func (o OutputOptions) limitNotes(notes commitNoteDetails) commitNoteDetails {
	ns := notes
	if o.Limit > 0 && len(ns) > o.Limit {
		ns = ns[0:o.Limit]
	}
	return ns
}

// Status returns the status report
func Status(n note.CommitNote, options OutputOptions, projPath ...string) (string, error) {
	if options.TerminalOff {
		n = n.FilterOutTerminal()
	}

	if options.TotalOnly {
		return util.DurationStr(n.Total()), nil
	}

	projName := ""
	tags := ""
	if len(projPath) > 0 {
		projName = filepath.Base(projPath[0])
		tagList, err := project.LoadTags(filepath.Join(projPath[0], ".gtm"))
		if err != nil {
			return "", err
		}
		tags = strings.Join(tagList, ",")
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Status").Funcs(funcMap).Parse(statusTpl))

	err := t.Execute(
		b,
		struct {
			ProjPath    []string
			ProjectName string
			commitNoteDetail
			BoldFormat string
			Tags       string
		}{
			projPath,
			projName,
			commitNoteDetail{Note: n},
			setBoldFormat(options.Color),
			tags,
		})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// Commits returns the commits report
func Commits(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))

	b := new(bytes.Buffer)
	t := template.Must(template.New("Commits").Funcs(funcMap).Parse(commitsTpl))

	err := t.Execute(
		b,
		struct {
			FullMessage bool
			Notes       commitNoteDetails
			BoldFormat  string
		}{
			options.FullMessage,
			notes,
			setBoldFormat(options.Color)})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Timeline returns the time spent by hour
func Timeline(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))
	timeline, err := notes.timeline()

	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineTpl))
	err = t.Execute(
		b,
		struct {
			Timeline   timelineEntries
			BoldFormat string
		}{
			timeline,
			setBoldFormat(options.Color),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// TimelineCommits returns the number commits by hour
func TimelineCommits(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))
	timeline, err := notes.timelineCommits()

	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineCommitTpl))
	err = t.Execute(
		b,
		struct {
			Timeline   timelineCommitEntries
			BoldFormat string
		}{
			timeline,
			setBoldFormat(options.Color),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Files returns the files report
func Files(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))

	b := new(bytes.Buffer)
	t := template.Must(template.New("Files").Funcs(funcMap).Parse(filesTpl))

	err := t.Execute(
		b,
		struct {
			Files fileEntries
		}{
			notes.files(),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil

}

func setBoldFormat(color bool) string {
	if (color || isatty.IsTerminal(os.Stdout.Fd())) && runtime.GOOS != "windows" {
		return "\x1b[1m%s\x1b[0m"
	}
	return "%s"
}

// BlockForVal determines the correct block to return for a value
func BlockForVal(val, max int) string {
	const (
		blockCnt   int = 8
		blockWidth int = 3
	)

	blocks := []string{`▁`, `▂`, `▃`, `▄`, `▅`, `▆`, `▇`, `█`}

	if max < blockCnt {
		max = blockCnt
	}

	if val == 0 {
		return strings.Repeat(" ", blockWidth)
	}

	inc := max / blockCnt
	if inc == 0 {
		return strings.Repeat(" ", blockWidth)
	}

	idx := val / inc

	// let's force to start at index 0
	if val == 1 && max == blockCnt {
		idx = 0
	}

	// let's make sure we don't get index out range panic
	if idx > blockCnt-1 {
		idx = blockCnt - 1
	}

	return strings.Repeat(blocks[idx], blockWidth)
}
