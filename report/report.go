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
{{- range $note := .Notes }}
	{{- $total := .Note.Total }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ printf $headerFormat $note.Project }} {{ $note.Author }}{{- printf "\n" }}
	{{- range $i, $f := .Note.Files }}
		{{- if $f.IsTerminal }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%5.2f"}}%% [{{ $f.Status }}] Terminal
		{{- else }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%5.2f"}}%% [{{ $f.Status }}] {{$f.SourceFile}}
		{{- end }}
	{{- end }}
	{{- if len .Note.Files }}
		{{- FormatDuration $total | printf "\n%14s\n\n" }}
	{{- else }}
		{{- printf "\n" }}
	{{- end }}
{{- end -}}`
	statusTpl string = `
{{- $headerFormat := .HeaderFormat }}
{{- if .Note.Files }}{{ printf "\n"}}{{end}}
{{- range $i, $f := .Note.Files }}
	{{- if $f.IsTerminal }}
		{{- FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] Terminal
	{{- else }}
		{{- FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
	{{- end }}
{{ end }}
{{- if len .Note.Files }}
	{{- FormatDuration .Note.Total | printf "%14s" }}      {{ printf $headerFormat .ProjectName }}
{{ end }}`
	commitTotalsTpl string = `
{{ $headerFormat := .HeaderFormat }}
{{- $total := .Notes.Total }}
{{- range $_, $note := .Notes }}
	{{- print "\n" }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ printf $headerFormat $note.Project }} {{ $note.Author }}{{- printf "\n" }}
	{{- if len .Note.Files }}
		{{- printf "\n" }}  {{ FormatDuration .Note.Total | printf "%14s" }}   {{ Percent $note.Note.Total $total | printf "%.2f"}}%%{{- print "\n" }}
	{{- end }}
{{- end }}
  {{ FormatDuration $total | printf "%14s" }} Total Hours {{ printf "\n" }}
`
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
		{{- $f.Duration | printf "%14s" }} {{ Percent $f.Seconds $total | printf "%5.2f"}}%%  Terminal
	{{- else }}
		{{- $f.Duration | printf "%14s" }} {{ Percent $f.Seconds $total | printf "%5.2f"}}%%  {{ $f.Filename }}
	{{- end }}
{{ end }}
{{- if len .Files }}
	{{- .Files.Duration | printf "%14s" }}
{{ end }}`
)

// Status returns the status report
func Status(n note.CommitNote, totalOnly bool, projPath ...string) (string, error) {
	if totalOnly {
		return util.DurationStr(n.Total()), nil
	}

	projName := ""
	if len(projPath) > 0 {
		projName = filepath.Base(projPath[0])
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("Status").Funcs(funcMap).Parse(statusTpl))

	headerFormat := "%s"
	if isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows" {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}

	err := t.Execute(
		b,
		struct {
			ProjectName string
			commitNoteDetail
			HeaderFormat string
		}{
			projName,
			commitNoteDetail{Note: n},
			headerFormat,
		})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// Commits returns the commits report
func Commits(projects []ProjectCommits, totalOnly bool, limit int) (string, error) {
	notes := retrieveNotes(projects)
	b := new(bytes.Buffer)
	var t *template.Template
	if totalOnly {
		t = template.Must(template.New("Commit Totals").Funcs(funcMap).Parse(commitTotalsTpl))
	} else {
		t = template.Must(template.New("Commits").Funcs(funcMap).Parse(commitsTpl))
	}
	headerFormat := "%s"
	if isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows" {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}

	if limit > 0 && len(notes) > limit {
		notes = notes[0:limit]
	}

	err := t.Execute(
		b,
		struct {
			Notes        commitNoteDetails
			HeaderFormat string
		}{
			notes,
			headerFormat})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Timeline returns the timeline report
func Timeline(projects []ProjectCommits, limit int) (string, error) {
	notes := retrieveNotes(projects)
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
func Files(projects []ProjectCommits, limit int) (string, error) {
	notes := retrieveNotes(projects)
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
