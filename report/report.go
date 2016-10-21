package report

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/git-time-metric/gtm/note"
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

type ProjectCommits struct {
	Path    string
	Commits []string
}

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

const (
	commitsTpl string = `
{{ $boldFormat := .BoldFormat }}
{{- $fullMessage := .FullMessage }}
{{- range $note := .Notes }}
	{{- $total := .Note.Total }}
	{{- printf $boldFormat $note.Hash }} {{ printf $boldFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ printf $boldFormat $note.Project }} {{ $note.Author }}{{- printf "\n" }}
	{{- if $fullMessage}}{{- if $note.Message }}{{- printf "\n"}}{{- $note.Message }}{{- printf "\n"}}{{end}}{{end}}
	{{- range $i, $f := .Note.Files }}
		{{- if $f.IsTerminal }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] Terminal
		{{- else }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] {{$f.SourceFile}}
		{{- end }}
	{{- end }}
	{{- if len .Note.Files }}
		{{- FormatDuration $total | printf "\n%14s" }}          {{ printf $boldFormat $note.Project }}{{ printf "\n\n" }}
	{{- else }}
		{{- printf "\n" }}
	{{- end }}
{{- end -}}`

	statusTpl string = `
{{- $boldFormat := .BoldFormat }}
{{- if .Note.Files }}{{ printf "\n"}}{{end}}
{{- $total := .Note.Total }}
{{- range $i, $f := .Note.Files }}
	{{- if $f.IsTerminal }}
		{{- FormatDuration $f.TimeSpent | printf "%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] Terminal
	{{- else }}
		{{- FormatDuration $f.TimeSpent | printf "%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] {{$f.SourceFile}}
	{{- end }}
{{ end }}
{{- if len .Note.Files }}
	{{- FormatDuration .Note.Total | printf "%14s" }}          {{ printf $boldFormat .ProjectName }}
{{ end }}`

	// TODO: determine left padding based on size of total duration
	timelineTpl string = `
{{- $boldFormat := .BoldFormat }}
{{- $maxSecondsInHour := .Timeline.HourMaxSeconds }}
{{printf $boldFormat "             00.01.02.03.04.05.06.07.08.09.10.11.12.01.02.03.04.05.06.07.08.09.10.11." }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ range $_, $entry := .Timeline }}
{{- printf $boldFormat $entry.Day }} | {{ range $_, $h := .Hours }}{{ Blocks $h $maxSecondsInHour }}{{ end }} | {{ LeftPad2Len $entry.Duration " " 13 | printf $boldFormat }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ end }}
{{- if len .Timeline }}
	{{- LeftPad2Len .Timeline.Duration " " 101 | printf $boldFormat }}
{{ end }}`

	// TODO: determine left padding based on total hours
	filesTpl string = `
{{- $total := .Files.Total }}
{{ range $i, $f := .Files }}
	{{- if $f.IsTerminal }}
		{{- $f.Duration | printf "%14s" }} {{ Percent $f.Seconds $total | printf "%3.0f"}}%%  Terminal
	{{- else }}
		{{- $f.Duration | printf "%14s" }} {{ Percent $f.Seconds $total | printf "%3.0f"}}%%  {{ $f.Filename }}
	{{- end }}
{{ end }}
{{- if len .Files }}
	{{- .Files.Duration | printf "%14s" }}
{{ end }}`
)

// Status returns the status report
func Status(n note.CommitNote, options OutputOptions, projPath ...string) (string, error) {
	if options.TerminalOff {
		n = n.FilterOutTerminal()
	}

	if options.TotalOnly {
		return util.DurationStr(n.Total()), nil
	}

	projName := ""
	if len(projPath) > 0 {
		projName = filepath.Base(projPath[0])
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Status").Funcs(funcMap).Parse(statusTpl))

	err := t.Execute(
		b,
		struct {
			ProjectName string
			commitNoteDetail
			BoldFormat string
		}{
			projName,
			commitNoteDetail{Note: n},
			setBoldFormat(options.Color),
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

// TODO: optional timeline that reports on total commits per hour
// Timeline returns the timeline report
func Timeline(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))
	timeline, err := notes.timeline()
	if err != nil {
		return "", err
	}

	// TODO: option to report on all days and not just days with commits
	// TODO: calculate average total daily hours
	// TODO: calculate busiest days of the week
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

func BlockForVal(val, max int) string {
	const (
		blockCnt   int = 8
		blockWidth int = 3
	)

	blocks := []string{`▁`, `▂`, `▃`, `▄`, `▅`, `▆`, `▇`, `█`}

	if val == 0 {
		return strings.Repeat(" ", blockWidth)
	}

	inc := max / blockCnt
	if inc == 0 {
		return strings.Repeat(" ", blockWidth)
	}

	// let's make sure we don't get index out range panic
	idx := val / inc
	if idx > blockCnt-1 {
		idx = blockCnt - 1
	}

	return strings.Repeat(blocks[idx], blockWidth)
}
