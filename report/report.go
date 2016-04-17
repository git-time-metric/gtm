package report

import (
	"bytes"
	"text/template"

	"edgeg.io/gtm/commit"
	"edgeg.io/gtm/util"
)

var funcMap = template.FuncMap{
	"left2PadLen":    util.LeftPad2Len,
	"formatDuration": util.FormatDuration,
}

const (
	FilesTpl string = `
{{ define "Files" -}}
{{ $ln := .Log.MaxSourceFileLen }}
{{ range $i, $f := .Log.Files -}}
{{   left2PadLen $f.SourceFile " " $ln }}: {{ formatDuration $f.TimeSpent | printf "%15s" }} [{{ $f.Status }}]
{{ end -}}
{{    if $ln -}}
{{       left2PadLen "Total" " " $ln }}: {{ formatDuration .Log.Total | printf "%15s" }}
{{    end -}}
{{ end }}
`
	MessageFilesTpl string = `
{{ range $_, $log := . }}
{{   $log.Message }}
{{-  template "Files" $log -}} 
{{ end -}}
`
	MessageTpl string = `
{{ range $_, $log := . -}}
{{-  $log.Message -}}
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
