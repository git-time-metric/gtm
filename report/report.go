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
	FilesTpl string = `
{{ define "Files" -}}
{{ range $i, $f := .Log.Files -}}
{{   FormatDuration $f.TimeSpent | printf "%14s" }}  [{{ $f.Status }}] {{$f.SourceFile}} 
{{ end -}}
{{    if len .Log.Files -}}
{{       FormatDuration .Log.Total | printf "%14s" }} 
{{    end -}}
{{ end }}
`
	MessageFilesTpl string = `
{{ range $_, $log := . }}
{{   $log.Message }}
{{   template "Files" $log }} 
{{ end -}}
`
)

func Files(l commit.Log) (string, error) {
	b := new(bytes.Buffer)
	t := template.Must(template.New("Files").Funcs(funcMap).Parse(FilesTpl))
	err := t.Execute(b, struct{ Log *commit.Log }{&l})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func MessageFiles(commits []string) (string, error) {
	logs, err := retrieveLogs(commits)
	if err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("Message Files").Funcs(funcMap).Parse(FilesTpl))
	t = template.Must(t.Parse(MessageFilesTpl))
	err = t.Execute(b, &logs)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
