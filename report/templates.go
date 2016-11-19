// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package report

const (
	commitsTpl string = `
{{ $boldFormat := .BoldFormat }}
{{- $fullMessage := .FullMessage }}
{{- range $note := .Notes }}
	{{- $total := .Note.Total }}
	{{- printf $boldFormat $note.Hash }} {{ printf $boldFormat $note.Subject }}{{- printf "\n" }}
	{{- $note.Date }} {{ printf $boldFormat $note.Project }} {{ $note.Author }}{{- printf "\n" }}
	{{- if $fullMessage}}{{- if $note.Message }}{{- printf "\n"}}{{- $note.Message }}{{- printf "\n"}}{{end}}{{end}}
	{{- range $i, $f := .Note.Files }}
		{{- if $f.IsTerminal }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] Terminal
		{{- else }}
			{{- FormatDuration $f.TimeSpent | printf "\n%14s" }} {{ Percent $f.TimeSpent $total | printf "%3.0f"}}%% [{{ $f.Status }}] {{$f.SourceFile}}
		{{- end }}
	{{- end }}
	{{- if len .Note.Files }}
		{{- FormatDuration $total | printf "\n%14s" }}          {{ printf $boldFormat $note.Project }}{{ printf "\n\n" }}
	{{- else }}
		{{- printf "\n" }}
	{{- end }}
{{- end -}}`

	statusTpl string = `
{{- $boldFormat := .BoldFormat }}
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
	{{- FormatDuration .Note.Total | printf "%14s" }}          {{ printf $boldFormat .ProjectName }} {{ if .Tags }}[{{ .Tags }}]{{ end }}
{{ end }}`

	// TODO: determine left padding based on size of total duration
	timelineTpl string = `
{{- $boldFormat := .BoldFormat }}
{{- $maxSecondsInHour := .Timeline.HourMaxSeconds }}
{{printf $boldFormat "             00.01.02.03.04.05.06.07.08.09.10.11.12.01.02.03.04.05.06.07.08.09.10.11." }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ range $_, $entry := .Timeline }}
{{- printf $boldFormat $entry.Day }} | {{ range $_, $h := .Hours }}{{ Blocks $h $maxSecondsInHour }}{{ end }} | {{ LeftPad2Len $entry.Duration " " 13 | printf $boldFormat }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ end }}
{{- if len .Timeline }}
	{{- LeftPad2Len .Timeline.Duration " " 101 | printf $boldFormat }}
{{ end }}`

	timelineCommitTpl string = `
{{- $boldFormat := .BoldFormat }}
{{- $maxCommitsInHour := .Timeline.HourMaxCommits}}
{{printf $boldFormat "             00.01.02.03.04.05.06.07.08.09.10.11.12.01.02.03.04.05.06.07.08.09.10.11." }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ range $_, $entry := .Timeline }}
{{- printf $boldFormat $entry.Day }} | {{ range $_, $c := .Commits }}{{ Blocks $c $maxCommitsInHour }}{{ end }} | {{ printf "%4d" $entry.Total | printf $boldFormat }}
{{printf $boldFormat "             ------------------------------------------------------------------------"}}
{{ end }}
{{- if len .Timeline }}
	{{- printf "%92d" .Timeline.Total | printf $boldFormat }}
{{ end }}`

	// TODO: determine left padding based on total hours
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
