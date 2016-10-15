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
	color := reportFlags.Bool(
		"color",
		false,
		"Always output color even if no terminal is detected, i.e 'gtm report -color | less -R'")
	terminalOff := reportFlags.Bool(
		"terminal-off",
		false,
		"Exclude time spent in terminal (Terminal plugin is required)")
	format := reportFlags.String(
		"format",
		"commits",
		"Specify report format [commits|files|timeline]")
	limit := reportFlags.Int(
		"n",
		0,
		fmt.Sprintf("Limit output, 0 is no limits, defaults to 1 when no limiting flags (i.e. -today, -author, etc) otherwise defaults to 0"))
	fullMessage := reportFlags.Bool(
		"full-message",
		false,
		"Include full commit message")
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
		"this-week",
		false,
		"Show commits for this week")
	lastWeek := reportFlags.Bool(
		"last-week",
		false,
		"Show commits for last week")
	thisMonth := reportFlags.Bool(
		"this-month",
		false,
		"Show commits for this month")
	lastMonth := reportFlags.Bool(
		"last-month",
		false,
		"Show commits for last month")
	thisYear := reportFlags.Bool(
		"this-year",
		false,
		"Show commits for this year")
	lastYear := reportFlags.Bool(
		"last-year",
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
		"Project tags to report on, i.e --tags tag1,tag2")
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

		tagList := []string{}
		if *tags != "" {
			tagList = util.Map(strings.Split(*tags, ","), strings.TrimSpace)
		}
		projects, err := index.Get(tagList, *all)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		limiter, err := scm.NewCommitLimiter(
			*limit, *before, *after, *author, *message,
			*today, *yesterday, *thisWeek, *lastWeek,
			*thisMonth, *lastMonth, *thisYear, *lastYear)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		*limit = limiter.Max

		for _, p := range projects {
			commits, err = scm.CommitIDs(limiter, p)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			projCommits = append(projCommits, report.ProjectCommits{Path: p, Commits: commits})
		}
	}

	options := report.OutputOptions{
		FullMessage: *fullMessage,
		TerminalOff: *terminalOff,
		Color:       *color,
		Limit:       *limit}

	switch *format {
	case "commits":
		out, err = report.Commits(projCommits, options)
	case "files":
		out, err = report.Files(projCommits, options)
	case "timeline":
		out, err = report.Timeline(projCommits, options)
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
