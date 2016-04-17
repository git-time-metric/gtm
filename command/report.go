package command

import (
	"bufio"
	"fmt"
	"os"

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
	gtm report commit-id ...

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
			fmt.Println("Unable to report, commit identifiers not provided")
			return 1
		}
		commits = args
	}
	out, err := report.MessageFiles(commits)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(out)
	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	`
}
