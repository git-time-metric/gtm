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
}

const (
	commitFilesTpl string = `
{{ define "Files" }}
{{ range $i, $f := .Note.Files -}}
	{{- FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
{{ end -}}
{{ if len .Note.Files -}}
	{{- FormatDuration .Note.Total | printf "%14s" }}
{{ end -}}
{{ end -}}
`
	commitDetailsTpl string = `
{{ $headerFormat := .HeaderFormat -}}
{{ range $_, $note := .Notes -}}
	{{- printf $headerFormat $note.Hash }} {{ printf $headerFormat $note.Subject }}
    {{- $note.Date }} {{ $note.Author }}
	{{ template "Files" $note }}
{{ end -}}
`
)

func NoteFiles(n note.CommitNote) (string, error) {
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))
	t = template.Must(t.Parse(`{{ template "Files" . }}`))

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
	notes, err := retrieveNotes(commits)
	if err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))
	t = template.Must(t.Parse(commitDetailsTpl))
	headerFormat := "%s"
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}
	err = t.Execute(
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
	notes, err := retrieveNotes(commits)
	if err != nil {
		return "", err
	}
	return util.FormatDuration(notes.Total()), nil
}
