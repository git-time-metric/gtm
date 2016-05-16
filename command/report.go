package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"edgeg.io/gtm/report"
	"edgeg.io/gtm/scm"
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
	reportFlags := flag.NewFlagSet("report", flag.ExitOnError)
	format := reportFlags.String(
		"format",
		"all",
		"Specify report format [all|total]")
	limit := reportFlags.Int(
		"n",
		1,
		fmt.Sprintf("Limit number of log enteries"))
	reportFlags.Parse(os.Args[2:])

	if !util.StringInSlice([]string{"all", "total"}, *format) {
		fmt.Printf("report --format=%s not valid\n", *format)
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			commits = append(commits, scanner.Text())
		}
	} else {
		if len(reportFlags.Args()) == 0 {
			commits, err = scm.GitLogSHA1s([]string{"-n", strconv.Itoa(*limit)})
			if err != nil {
				fmt.Println(err)
				return 1
			}
		}
		for _, a := range reportFlags.Args() {
			if match, err := regexp.MatchString("[-|.|,|:|*]", a); err != nil || match {
				fmt.Printf("\nNot a valid commit sha1 %s\n", a)
				return 1
			}
			commits = append(commits, a)
		}
	}

	if len(commits) == 0 {
		return 0
	}

	if *format == "total" {
		out, err = report.NoteDetailsTotal(commits)
	} else {
		out, err = report.NoteDetails(commits)
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
	Show commit time logs
	`
}
