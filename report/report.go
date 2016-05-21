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
	commitDetailsTpl string = `
{{ $headerFormat := .HeaderFormat }}
{{- range $_, $note := .Notes }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}
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
	commitFilesTpl string = `
{{ range $i, $f := .Note.Files }}
	{{- FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
{{ end }}
{{- if len .Note.Files }}
	{{- FormatDuration .Note.Total | printf "%14s" }}
{{ end }}`
	commitTotalsTpl string = `
{{ $headerFormat := .HeaderFormat }}
{{- range $_, $note := .Notes }}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}
	{{- $note.Date }} {{ $note.Author }}  {{if len .Note.Files }}{{ FormatDuration .Note.Total }}{{ end }}
	{{- print "\n" }}
{{ end }}`
	timelineTpl string = `
           0123456789012345678901234
{{ range $_, $entry := .Timeline }}
	{{- $entry.Day }} {{ RightPad2Len $entry.Bars " " 24 }} {{ LeftPad2Len $entry.Duration " " 13 }}
{{ end }}
{{- if len .Timeline }}
{{- LeftPad2Len "------------" " " 49 }}
{{ LeftPad2Len .Timeline.Duration " " 49 }}
{{ end }}`
)

func NoteFiles(n note.CommitNote) (string, error) {
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))

	err := t.Execute(b, commitNoteDetail{Note: n})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func NoteFilesTotal(n note.CommitNote) string {
	return util.FormatDuration(n.Total())
}

func NoteDetails(commits []string) (string, error) {
	notes := retrieveNotes(commits)
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitDetailsTpl))
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

func NoteDetailsTotal(commits []string) (string, error) {
	notes := retrieveNotes(commits)
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Totals").Funcs(funcMap).Parse(commitTotalsTpl))
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
			Timeline timeline
		}{
			notes.Timeline(),
		})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
