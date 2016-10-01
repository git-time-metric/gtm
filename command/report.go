package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/report"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
)

type ReportCmd struct {
}

func NewReport() (cli.Command, error) {
	return ReportCmd{}, nil
}

func (r ReportCmd) Help() string {
	return r.Synopsis()
}

func (r ReportCmd) Run(args []string) int {
	reportFlags := flag.NewFlagSet("report", flag.ExitOnError)
	format := reportFlags.String(
		"format",
		"commits",
		"Specify report format [commits|files|timeline]")
	limit := reportFlags.Int(
		"n",
		0,
		fmt.Sprintf("Limit number of log enteries"))
	totalOnly := reportFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	before := reportFlags.String(
		"before",
		"",
		"Show commits older than a specific date")
	after := reportFlags.String(
		"after",
		"",
		"Show commits more recent than a specific date")
	today := reportFlags.Bool(
		"today",
		false,
		"Show commits for today")
	yesterday := reportFlags.Bool(
		"yesterday",
		false,
		"Show commits for yesterday")
	thisWeek := reportFlags.Bool(
		"thisweek",
		false,
		"Show commits for this week")
	lastWeek := reportFlags.Bool(
		"lastweek",
		false,
		"Show commits for last week")
	thisMonth := reportFlags.Bool(
		"thismonth",
		false,
		"Show commits for this month")
	lastMonth := reportFlags.Bool(
		"lastmonth",
		false,
		"Show commits for last month")
	thisYear := reportFlags.Bool(
		"thisyear",
		false,
		"Show commits for this year")
	lastYear := reportFlags.Bool(
		"lastyear",
		false,
		"Show commits for last year")
	author := reportFlags.String(
		"author",
		"",
		"Show commits which contain author substring")
	message := reportFlags.String(
		"message",
		"",
		"Show commits which contain message substring")
	tags := reportFlags.String(
		"tags",
		"",
		"Project tags to report on")
	all := reportFlags.Bool(
		"all",
		false,
		"Show commits for all projects")

	if err := reportFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if !util.StringInSlice([]string{"commits", "timeline", "files"}, *format) {
		fmt.Fprintf(os.Stderr, "report --format=%s not valid\n", *format)
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	const (
		invalidSHA1 = "\nNot a valid commit SHA-1 %s\n"
	)

	// if running from within a MINGW console isatty detection does not work
	// https://github.com/mintty/mintty/issues/482
	isMinGW := strings.HasPrefix(os.Getenv("MSYSTEM"), "MINGW")

	sha1Regex := regexp.MustCompile(`\A([0-9a-f]{40})\z`)

	projCommits := []report.ProjectCommits{}

	switch {
	case !isMinGW && !isatty.IsTerminal(os.Stdin.Fd()):
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if !sha1Regex.MatchString(scanner.Text()) {
				fmt.Fprintf(os.Stderr, invalidSHA1, scanner.Text())
				return 1
			}
			commits = append(commits, scanner.Text())
		}
		curProjPath, err := scm.RootPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		projCommits = append(projCommits, report.ProjectCommits{Path: curProjPath, Commits: commits})

	case len(reportFlags.Args()) > 0:
		for _, a := range reportFlags.Args() {
			if !sha1Regex.MatchString(a) {
				fmt.Fprintf(os.Stderr, invalidSHA1, a)
				return 1
			}
			commits = append(commits, a)
		}
		curProjPath, err := scm.RootPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		projCommits = append(projCommits, report.ProjectCommits{Path: curProjPath, Commits: commits})

	default:
		index, err := project.NewIndex()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		projects, err := index.Get(strings.Fields(*tags), *all)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		limiter, err := scm.NewCommitLimiter(
			*limit, *before, *after, *author, *message,
			*today, *yesterday, *thisWeek, *lastWeek,
			*thisMonth, *lastMonth, *thisYear, *lastYear)

		for _, p := range projects {
			commits, err = scm.CommitIDs(limiter, p)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			projCommits = append(projCommits, report.ProjectCommits{Path: p, Commits: commits})
		}
	}

	// if len(projCommits) == 0 || len(projCommits[0].Commits) == 0 {
	// 	return 0
	// }

	switch *format {
	case "commits":
		out, err = report.Commits(projCommits, *totalOnly, *limit)
	case "files":
		out, err = report.Files(projCommits, *limit)
	case "timeline":
		out, err = report.Timeline(projCommits, *limit)
	case "projects":
	case "json":
	case "csv":
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf(out)

	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	Usage: gtm report [-n] [-format commits|files|timeline] [-total-only]
	Report on time logged
	`
}
