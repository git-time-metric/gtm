// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
)

// VerifyCmd contains CLI commands for verify
type VerifyCmd struct {
	UI      cli.Ui
	Version string
	Out     *bytes.Buffer
}

func (c VerifyCmd) output(s string) {
	var err error
	if c.Out != nil {
		_, err = fmt.Fprint(c.Out, s)
	} else {
		_, err = fmt.Fprint(os.Stdout, s)
	}
	if err != nil {
		fmt.Printf("Error printing output, %s\n", err)
	}
}

// Help returns CLI help for Verify command
func (c VerifyCmd) Help() string {
	helpText := `
Usage: gtm verify <version-constraint>

  Check if gtm satisfies a Semantic Version 2.0 constraint.
`
	return strings.TrimSpace(helpText)
}

// Run executes verify commands with args
func (c VerifyCmd) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("verify", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(args) == 0 {
		c.UI.Error("Unable to verify version, version constraint not provided")
		return 1
	}

	valid, err := c.check(args[0])
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	c.output(fmt.Sprintf("%t", valid))
	return 0
}

// Synopsis returns verify help
func (c VerifyCmd) Synopsis() string {
	return "Check if gtm satisfies a Semantic Version 2.0 constraint"
}

func (c VerifyCmd) check(constraint string) (bool, error) {
	// Our version tags can have a 'v' prefix
	// Strip v prefix if it exists because it's not valid for a Semantic version
	cleanVersion := c.Version
	if strings.HasPrefix(strings.ToLower(cleanVersion), "v") {
		cleanVersion = cleanVersion[1:]
	}

	ver, err := version.NewVersion(cleanVersion)
	if err != nil {
		return false, err
	}

	vc, err := version.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	return vc.Check(ver), nil
}
