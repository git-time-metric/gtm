// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
	isatty "github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
)

// ReportCmd contains methods for report command
type ReportCmd struct {
	UI cli.Ui
}

// NewReport create new ReportCmd struct
func NewReport() (cli.Command, error) {
	return ReportCmd{}, nil
}

// Help returns help for report command
func (c ReportCmd) Help() string {
	helpText := `
Usage: gtm report [options] <Commit-ID>...

  Display reports for one or more git repositories.

Options:

  Report Formats:

  -format=commits            Specify report format [summary|project|commits|files|timeline-hours|timeline-commits] (default commits)
  -full-message=false        Include full commit message
  -terminal-off=false        Exclude time spent in terminal (Terminal plug-in is required)
  -app-off=false             Exclude time spent in apps
  -force-color=false         Always output color even if no terminal is detected, i.e 'gtm report -color | less -R'
  -testing=false             This is used for automated testing to force default test path

  Commit Limiting:

  -n int=1                   Limit output, 0 is no limits, defaults to 1 when no limiting flags otherwise defaults to 0
  -from-date=yyyy-mm-dd      Show commits starting from this date
  -to-date=yyyy-mm-dd        Show commits thru the end of this date
  -author=""                 Show commits which contain author substring
  -message=""                Show commits which contain message substring
  -today=false               Show commits for today
  -yesterday=false           Show commits for yesterday
  -this-week=false           Show commits for this week
  -last-week=false           Show commits for last week
  -this-month=false          Show commits for this month
  -last-month=false          Show commits for last month
  -this-year=false           Show commits for this year
  -last-year=false           Show commits for last year

  Multi-Project Reporting:

  -tags=""                   Project tags to report on, i.e --tags tag1,tag2
  -all=false                 Show commits for all projects
`
	return strings.TrimSpace(helpText)
}

// Run executes report command with args
func (c ReportCmd) Run(args []string) int {
	var limit int
	var color, terminalOff, appOff, fullMessage, testing bool
	var today, yesterday, thisWeek, lastWeek, thisMonth, lastMonth, thisYear, lastYear, all bool
	var fromDate, toDate, message, author, tags, format string
	cmdFlags := flag.NewFlagSet("report", flag.ContinueOnError)
	cmdFlags.BoolVar(&color, "force-color", false, "")
	cmdFlags.BoolVar(&terminalOff, "terminal-off", false, "")
	cmdFlags.BoolVar(&appOff, "app-off", false, "")
	cmdFlags.StringVar(&format, "format", "commits", "")
	cmdFlags.IntVar(&limit, "n", 0, "")
	cmdFlags.BoolVar(&fullMessage, "full-message", false, "")
	cmdFlags.StringVar(&fromDate, "from-date", "", "")
	cmdFlags.StringVar(&toDate, "to-date", "", "")
	cmdFlags.BoolVar(&today, "today", false, "")
	cmdFlags.BoolVar(&yesterday, "yesterday", false, "")
	cmdFlags.BoolVar(&thisWeek, "this-week", false, "")
	cmdFlags.BoolVar(&lastWeek, "last-week", false, "")
	cmdFlags.BoolVar(&thisMonth, "this-month", false, "")
	cmdFlags.BoolVar(&lastMonth, "last-month", false, "")
	cmdFlags.BoolVar(&thisYear, "this-year", false, "")
	cmdFlags.BoolVar(&lastYear, "last-year", false, "")
	cmdFlags.StringVar(&author, "author", "", "")
	cmdFlags.StringVar(&message, "message", "", "")
	cmdFlags.StringVar(&tags, "tags", "", "")
	cmdFlags.BoolVar(&all, "all", false, "")
	cmdFlags.BoolVar(&testing, "testing", false, "")
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !util.StringInSlice([]string{"summary", "commits", "timeline-hours", "files", "timeline-commits", "project"}, format) {
		c.UI.Error(fmt.Sprintf("report --format=%s not valid\n", format))
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	const invalidSHA1 = "\nNot a valid commit SHA-1 %s\n"

	// if running from within a MINGW console isatty detection does not work
	// https://github.com/mintty/mintty/issues/482
	isMinGW := strings.HasPrefix(os.Getenv("MSYSTEM"), "MINGW")

	sha1Regex := regexp.MustCompile(`\A([0-9a-f]{40})\z`)

	projCommits := []report.ProjectCommits{}

	switch {
	case !testing && !isMinGW && !isatty.IsTerminal(os.Stdin.Fd()):
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if !sha1Regex.MatchString(scanner.Text()) {
				c.UI.Error(fmt.Sprintf("%s %s", invalidSHA1, scanner.Text()))
				return 1
			}
			commits = append(commits, scanner.Text())
		}
		curProjPath, err := scm.GitRepoPath()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		curProjPath, err = scm.Workdir(curProjPath)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		projCommits = append(projCommits, report.ProjectCommits{Path: curProjPath, Commits: commits})

	case !testing && len(cmdFlags.Args()) > 0:
		for _, a := range cmdFlags.Args() {
			if !sha1Regex.MatchString(a) {
				c.UI.Error(fmt.Sprintf("%s %s", invalidSHA1, a))
				return 1
			}
			commits = append(commits, a)
		}
		curProjPath, err := scm.GitRepoPath()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		curProjPath, err = scm.Workdir(curProjPath)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		projCommits = append(projCommits, report.ProjectCommits{Path: curProjPath, Commits: commits})

	default:
		index, err := project.NewIndex()
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		tagList := []string{}
		if tags != "" {
			tagList = util.Map(strings.Split(tags, ","), strings.TrimSpace)
		}
		projects, err := index.Get(tagList, all)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		// hack, if project format we want all commits for the project
		if format == "project" && limit == 0 {
			// set max to absurdly high value for number of possible commits
			limit = 2147483647
		}

		limiter, err := scm.NewCommitLimiter(
			limit, fromDate, toDate, author, message,
			today, yesterday, thisWeek, lastWeek,
			thisMonth, lastMonth, thisYear, lastYear)

		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		limit = limiter.Max

		for _, p := range projects {
			commits, err = scm.CommitIDs(limiter, p)
			if err != nil {
				c.UI.Error(err.Error())
				return 1
			}
			projCommits = append(projCommits, report.ProjectCommits{Path: p, Commits: commits})
		}
	}

	options := report.OutputOptions{
		FullMessage: fullMessage,
		TerminalOff: terminalOff,
		AppOff:      appOff,
		Color:       color,
		Limit:       limit}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	switch format {
	case "project":
		out, err = report.ProjectSummary(projCommits, options)
	case "summary":
		out, err = report.CommitSummary(projCommits, options)
	case "commits":
		out, err = report.Commits(projCommits, options)
	case "files":
		out, err = report.Files(projCommits, options)
	case "timeline-hours":
		out, err = report.Timeline(projCommits, options)
	case "timeline-commits":
		out, err = report.TimelineCommits(projCommits, options)
	}

	s.Stop()

	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	c.UI.Output(out)

	return 0
}

// Synopsis return help for report command
func (c ReportCmd) Synopsis() string {
	return "Display reports for git repositories"
}
