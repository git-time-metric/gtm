// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"bytes"
	"fmt"
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
	TotalOnly    bool
	LongDuration bool
	FullMessage  bool
	TerminalOff  bool
	AppOff       bool
	Color        bool
	Limit        int
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
	defer util.Profile()()

	if options.TerminalOff {
		n = n.FilterOutTerminal()
	}
	if options.AppOff {
		n = n.FilterOutApp()
	}

	if options.TotalOnly {
		if options.LongDuration {
			return util.DurationStrLong(n.Total()), nil
		}
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
	cf := colorFormater{color: options.Color}
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
			cf.white(true),
			tags,
		})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// CommitSummary returns the commit summary report
func CommitSummary(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, false, "Mon Jan 02"))
	if len(notes) == 0 {
		return "", nil
	}

	lines := commitSummaryBuilder{}.Build(notes)

	b := new(bytes.Buffer)
	t := template.Must(template.New("Commits").Funcs(funcMap).Parse(commitSummaryTpl))
	cf := colorFormater{color: options.Color}
	err := t.Execute(
		b,
		struct {
			Lines       []commitSummaryLine
			BoldFormat  string
			GreenFormat string
		}{
			lines,
			cf.white(true),
			cf.green(false),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// ProjectSummary returns the project summary report
func ProjectSummary(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, false, "Mon Jan 02"))
	if len(notes) == 0 {
		return "", nil
	}

	projectTotals := map[string]int{}
	for _, n := range notes {
		projectTotals[n.Project] += n.Note.Total()
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("ProjectSummary").Funcs(funcMap).Parse(projectTotalsTpl))
	cf := colorFormater{color: options.Color}
	err := t.Execute(
		b,
		struct {
			Projects    map[string]int
			BoldFormat  string
			GreenFormat string
		}{
			projectTotals,
			cf.white(true),
			cf.green(false),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Commits returns the commits report
func Commits(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, true, ""))
	if len(notes) == 0 {
		return "", nil
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("CommitSummary").Funcs(funcMap).Parse(commitsTpl))
	cf := colorFormater{color: options.Color}
	err := t.Execute(
		b,
		struct {
			FullMessage bool
			Notes       commitNoteDetails
			BoldFormat  string
			GreenFormat string
		}{
			options.FullMessage,
			notes,
			cf.white(true),
			cf.green(false),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Timeline returns the time spent by hour
func Timeline(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, false, ""))
	if len(notes) == 0 {
		return "", nil
	}

	timeline, err := notes.timeline()

	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineTpl))
	cf := colorFormater{color: options.Color}
	err = t.Execute(
		b,
		struct {
			Timeline    timelineEntries
			BoldFormat  string
			GreenFormat string
		}{
			timeline,
			cf.white(true),
			cf.green(false),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// TimelineCommits returns the number commits by hour
func TimelineCommits(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, false, ""))
	if len(notes) == 0 {
		return "", nil
	}

	timeline, err := notes.timelineCommits()

	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineCommitTpl))
	cf := colorFormater{color: options.Color}
	err = t.Execute(
		b,
		struct {
			Timeline    timelineCommitEntries
			BoldFormat  string
			GreenFormat string
		}{
			timeline,
			cf.white(true),
			cf.green(false),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Files returns the files report
func Files(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff, options.AppOff, false, ""))
	if len(notes) == 0 {
		return "", nil
	}

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

type colorFormater struct {
	color bool
}

func (c colorFormater) hasColor() bool {
	return (c.color || isatty.IsTerminal(os.Stdout.Fd())) && runtime.GOOS != "windows"
}

func (c colorFormater) white(bold bool) string {
	var attrBold int
	if bold {
		attrBold = 1
	}
	if c.hasColor() {
		return fmt.Sprintf("\033[%d;%dm%%s\033[0m", attrBold, 97)
	}
	return "%s"
}

func (c colorFormater) green(bold bool) string {
	var attrBold int
	if bold {
		attrBold = 1
	}
	if c.hasColor() {
		return fmt.Sprintf("\033[%d;%dm%%s\033[0m", attrBold, 32)
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
