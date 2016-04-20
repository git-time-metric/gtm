package report

import (
	"bytes"
	"text/template"

	"edgeg.io/gtm/note"
	"edgeg.io/gtm/util"
)

var funcMap = template.FuncMap{
	"FormatDuration": util.FormatDuration,
}

const (
	commitFilesTpl string = `
{{ define "Files" -}}
{{ range $i, $f := .Note.Files -}}
{{   FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
{{ end -}}
{{    if len .Note.Files -}}
{{       FormatDuration .Note.Total | printf "%14s" }}
{{    end -}}
{{ end }}
`
	commitDetailsTpl string = `
{{ range $_, $log := . }}
{{   $log.Message }}
{{   template "Files" $log }}
{{ end -}}
`
)

func NoteFiles(n note.CommitNote) (string, error) {
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))
	t = template.Must(t.Parse("{{ template \"Files\" . }}"))

	err := t.Execute(b, commitNoteDetail{Note: n})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func NoteDetails(commits []string) (string, error) {
	logs, err := retrieveNotes(commits)
	if err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))
	t = template.Must(t.Parse(commitDetailsTpl))
	err = t.Execute(b, logs)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
