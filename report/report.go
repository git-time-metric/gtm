package report

import (
	"bytes"
	"text/template"

	"edgeg.io/gtm/commit"
	"edgeg.io/gtm/util"
)

var funcMap = template.FuncMap{
	"FormatDuration": util.FormatDuration,
}

const (
	commitFilesTpl string = `
{{ define "Files" -}}
{{ range $i, $f := .Log.Files -}}
{{   FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}}
{{ end -}}
{{    if len .Log.Files -}}
{{       FormatDuration .Log.Total | printf "%14s" }}
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

func CommitFiles(l commit.Log) (string, error) {
	b := new(bytes.Buffer)
	t := template.Must(template.New("Commit Details").Funcs(funcMap).Parse(commitFilesTpl))
	t = template.Must(t.Parse("{{ template \"Files\" . }}"))

	err := t.Execute(b, messageLog{Log: l})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func CommitDetails(commits []string) (string, error) {
	logs, err := retrieveLogs(commits)
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
