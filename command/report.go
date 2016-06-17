package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"

	"edgeg.io/gtm/report"
	"edgeg.io/gtm/scm"
	"edgeg.io/gtm/util"
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
		"Specify report format [commits|totals|files|timeline]")
	limit := reportFlags.Int(
		"n",
		0,
		fmt.Sprintf("Limit number of log enteries"))
	totalOnly := reportFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	if err := reportFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	if !util.StringInSlice([]string{"commits", "timeline", "files"}, *format) {
		fmt.Printf("report --format=%s not valid\n", *format)
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	sha1Regex := regexp.MustCompile(`\A([0-9a-f]{40})\z`)

	for _, a := range reportFlags.Args() {
		if !sha1Regex.MatchString(a) {
			fmt.Printf("\nNot a valid commit SHA %s\n", a)
			return 1
		}
		commits = append(commits, a)
	}

	// if running from within a MINGW console isatty does not work
	// https://github.com/mintty/mintty/issues/482
	if !isatty.IsTerminal(os.Stdin.Fd()) && len(commits) == 0 && *limit == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if !sha1Regex.MatchString(scanner.Text()) {
				fmt.Printf("\nNot a valid commit SHA %s\n", scanner.Text())
				return 1
			}
			commits = append(commits, scanner.Text())
		}
	} else {
		if len(commits) == 0 {
			if *limit == 0 {
				*limit = 1
			}
			commits, err = scm.CommitIDs(*limit)
			if err != nil {
				fmt.Println(err)
				return 1
			}
		}
	}

	if len(commits) == 0 {
		return 0
	}

	switch *format {
	case "commits":
		out, err = report.Commits(commits, *totalOnly)
	case "files":
		out, err = report.Files(commits)
	case "timeline":
		out, err = report.Timeline(commits)
	}

	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf(out)

	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	Usage: gtm report [-n] [-format commits|totals|files|timeline] [-total-only]
	Report on time logged
	`
}
