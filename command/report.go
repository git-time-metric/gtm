package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"edgeg.io/gtm/metric"
	"edgeg.io/gtm/report"
	"edgeg.io/gtm/util"
	"github.com/mitchellh/cli"
)

type ReportCmd struct {
}

func NewReport() (cli.Command, error) {
	return ReportCmd{}, nil
}

func (r ReportCmd) Help() string {
	return `
	Show commit time logs

	Show log for specific sha1 commits:
	gtm report sha1 ...

	Show log by piping output from git log:
	git report -1 --pretty=%H|gtm report
	`
}

func (r ReportCmd) Run(args []string) int {
	const (
		all   = "all"
		total = "total"
	)

	const (
		working   = "working"
		staging   = "staging"
		committed = "committed"
	)

	stateMap := map[string]metric.GitState{
		working:   metric.Working,
		staging:   metric.Staging,
		committed: metric.Committed,
	}

	var (
		formats = []string{all, total}
		states  = []string{working, staging, committed}
	)

	reportFlags := flag.NewFlagSet("report", flag.ExitOnError)
	format := reportFlags.String(
		"format",
		"all",
		fmt.Sprintf("Specify report format [%s]", strings.Join(formats, "|")))
	state := reportFlags.String(
		"state",
		"committed",
		fmt.Sprintf("Specify git status to report on [%s]", strings.Join(states, "|")))
	reportFlags.Parse(os.Args[2:])

	if !util.StringInSlice(formats, *format) {
		fmt.Printf("report --format=%s not valid\n", *format)
		return 1
	}

	if !util.StringInSlice(states, *state) {
		fmt.Printf("report --state=%s not valid\n", *state)
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	if *state == committed {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				commits = append(commits, scanner.Text())
			}
		} else {
			if len(args) == 0 {
				fmt.Println("Unable to show time log, commit sha1 not provided")
				return 1
			}
			for _, a := range args {
				if match, err := regexp.MatchString("[-|.|,|:|*]", a); err != nil || match {
					fmt.Printf("\nNot a valid commit sha1 %s\n", a)
					return 1
				}
				commits = append(commits, a)
			}
		}

		if *format == total {
			out, err = report.NoteTotal(commits)
		} else {
			out, err = report.NoteDetails(commits)
		}

		if err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Printf(out)

	} else {
		n, err := metric.Process(stateMap[*state], false)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		if *format == total {
			out, err = report.NoteFiles(n)
		} else {
			out, err = report.NoteFiles(n)
		}

		if err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Printf(out)

	}
	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	Show commit time logs
	`
}
