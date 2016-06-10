package report

import (
	"bytes"
	"os"
	"text/template"

	"golang.org/x/crypto/ssh/terminal"

	"edgeg.io/gtm/note"
	"edgeg.io/gtm/util"
)

var funcMap = template.FuncMap{
	"FormatDuration": util.FormatDuration,
	"RightPad2Len":   util.RightPad2Len,
	"LeftPad2Len":    util.LeftPad2Len,
}

const (
	commitsTpl string = `
{{ $headerFormat := .HeaderFormat }}
{{- range $_, $note := .Notes }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ $note.Author }} {{- printf "\n" }}
	{{- range $i, $f := .Note.Files }}
		{{- FormatDuration $f.TimeSpent | printf "\n%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
	{{- end }}
	{{- if len .Note.Files }}
		{{- FormatDuration .Note.Total | printf "\n%14s\n\n" }}
	{{- else }}
		{{- printf "\n" }}
	{{- end }}
{{- end -}}`
	statusTpl string = `
{{ range $i, $f := .Note.Files }}
	{{- FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
{{ end }}
{{- if len .Note.Files }}
	{{- FormatDuration .Note.Total | printf "%14s" }}
{{ end }}`
	commitTotalsTpl string = `
{{ $headerFormat := .HeaderFormat }}
{{- range $_, $note := .Notes }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ $note.Author }}  {{if len .Note.Files }}{{ FormatDuration .Note.Total }}{{ end }}
	{{- print "\n" }}
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
{{ range $i, $f := .Files }}
	{{- $f.Duration | printf "%14s" }}  {{ $f.Filename }}
{{ end }}
{{- if len .Files }}
	{{- .Files.Duration | printf "%14s" }}
{{ end }}`
)

func Status(n note.CommitNote, totalOnly bool) (string, error) {
	if totalOnly {
		return util.FormatDuration(n.Total()), nil
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("Status").Funcs(funcMap).Parse(statusTpl))

	err := t.Execute(b, commitNoteDetail{Note: n})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func Commits(commits []string, totalOnly bool) (string, error) {
	notes := retrieveNotes(commits)
	b := new(bytes.Buffer)
	var t *template.Template
	if totalOnly {
		t = template.Must(template.New("Commit Totals").Funcs(funcMap).Parse(commitTotalsTpl))
	} else {
		t = template.Must(template.New("Commits").Funcs(funcMap).Parse(commitsTpl))
	}
	headerFormat := "%s"
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		headerFormat = "\x1b[1m%s\x1b[0m"
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

func Timeline(commits []string) (string, error) {
	notes := retrieveNotes(commits)
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

func Files(commits []string) (string, error) {
	notes := retrieveNotes(commits)
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
