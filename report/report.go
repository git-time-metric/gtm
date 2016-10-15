package report

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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
{{ $headerFormat := .HeaderFormat }}
{{- $fullMessage := .FullMessage }}
{{- $totalOnly := .TotalOnly }}
{{- range $note := .Notes }}
	{{- $total := .Note.Total }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ printf $headerFormat $note.Project }} {{ $note.Author }}{{- printf "\n" }}
	{{- if $fullMessage}}{{- if $note.Message }}{{- printf "\n"}}{{- $note.Message }}{{- printf "\n"}}{{end}}{{end}}
	{{- if not $totalOnly }}
		{{- range $i, $f := .Note.Files }}
			{{- if $f.IsTerminal }}
				{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] Terminal
			{{- else }}
				{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] {{$f.SourceFile}}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if len .Note.Files }}
		{{- FormatDuration $total | printf "\n%14s" }}          {{ printf $headerFormat $note.Project }}{{ printf "\n\n" }}
	{{- else }}
		{{- printf "\n" }}
	{{- end }}
{{- end -}}`
	statusTpl string = `
{{- $headerFormat := .HeaderFormat }}
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
	{{- FormatDuration .Note.Total | printf "%14s" }}          {{ printf $headerFormat .ProjectName }}
{{ end }}`
	timelineTpl string = `
           0123456789012345678901234
{{ range $_, $entry := .Timeline }}
	{{- $entry.Day }} {{ RightPad2Len $entry.Bars " " 24 }} {{ LeftPad2Len $entry.Duration " " 13 }}
{{ end }}
{{- if len .Timeline }}
	{{- LeftPad2Len .Timeline.Duration " " 49 }}
{{ end }}`
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
			HeaderFormat string
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
			TotalOnly    bool
			FullMessage  bool
			Notes        commitNoteDetails
			HeaderFormat string
		}{
			options.TotalOnly,
			options.FullMessage,
			notes,
			setBoldFormat(options.Color)})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Timeline returns the timeline report
func Timeline(projects []ProjectCommits, options OutputOptions) (string, error) {
	notes := options.limitNotes(retrieveNotes(projects, options.TerminalOff))

	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineTpl))

	err := t.Execute(
		b,
		struct {
			Timeline timelineEntries
		}{
			notes.timeline(),
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
