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
func Status(n note.CommitNote, totalOnly, terminalOff, color bool, projPath ...string) (string, error) {
	if terminalOff {
		n = n.FilterOutTerminal()
	}

	if totalOnly {
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
			setBoldFormat(color),
		})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// Commits returns the commits report
func Commits(projects []ProjectCommits, totalOnly, fullMessage, terminalOff, color bool, limit int) (string, error) {
	notes := retrieveNotes(projects, terminalOff)

	b := new(bytes.Buffer)
	t := template.Must(template.New("Commits").Funcs(funcMap).Parse(commitsTpl))

	if limit > 0 && len(notes) > limit {
		notes = notes[0:limit]
	}

	err := t.Execute(
		b,
		struct {
			TotalOnly    bool
			FullMessage  bool
			Notes        commitNoteDetails
			HeaderFormat string
		}{
			totalOnly,
			fullMessage,
			notes,
			setBoldFormat(color)})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Timeline returns the timeline report
func Timeline(projects []ProjectCommits, terminalOff, color bool, limit int) (string, error) {
	notes := retrieveNotes(projects, terminalOff)
	b := new(bytes.Buffer)
	t := template.Must(template.New("Timeline").Funcs(funcMap).Parse(timelineTpl))

	if limit > 0 && len(notes) > limit {
		notes = notes[0:limit]
	}

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
func Files(projects []ProjectCommits, terminalOff, color bool, limit int) (string, error) {
	notes := retrieveNotes(projects, terminalOff)
	b := new(bytes.Buffer)
	t := template.Must(template.New("Files").Funcs(funcMap).Parse(filesTpl))

	if limit > 0 && len(notes) > limit {
		notes = notes[0:limit]
	}

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
