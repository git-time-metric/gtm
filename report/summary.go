package report

type commitSummaryLine struct {
	StartGroup bool
	EndGroup   bool
	CommitLine bool
	Date       string
	Subject    string
	Project    string
	Total      int
}

type commitSummaryBuilder struct {
}

func (c commitSummaryBuilder) Build(notes commitNoteDetails) []commitSummaryLine {
	total := 0
	lines := []commitSummaryLine{}
	for idx, n := range notes {
		if idx == 0 || notes[idx].Date != notes[idx-1].Date {
			if idx != 0 {
				lines = append(lines, commitSummaryLine{EndGroup: true, Total: total})
			}
			total = 0
			lines = append(lines, commitSummaryLine{StartGroup: true, Date: n.Date})
		}
		lines = append(lines, commitSummaryLine{CommitLine: true, Subject: n.Subject, Project: n.Project, Total: n.Note.Total()})
		total += n.Note.Total()
	}
	lines = append(lines, commitSummaryLine{EndGroup: true, Total: total})
	return lines
}
