package command

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"edgeg.io/gtm/report"

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
	var commits []string
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
	out, err := report.NoteDetails(commits)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(out)
	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	Show commit time logs
	`
}
