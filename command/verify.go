// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
)

// VerifyCmd contains CLI commands for verify
type VerifyCmd struct {
	Ui      cli.Ui
	Version string
}

// NewVerify returns new VerifyCmd struct with version set
func NewVerify(v string) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return VerifyCmd{Version: v}, nil
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
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(args) == 0 {
		c.Ui.Error("Unable to verify version, version constraint not provided")
		return 1
	}

	valid, err := c.check(args[0])
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	c.Ui.Output(fmt.Sprintf("%t", valid))
	return 0
}

// Synopsis returns verify help
func (c VerifyCmd) Synopsis() string {
	return "Check if gtm satisfies a Semantic Version 2.0 constraint"
}

func (v VerifyCmd) check(constraint string) (bool, error) {
	// Our version tags can have a 'v' prefix
	// Strip v prefix if it exists because it's not valid for a Semantic version
	cleanVersion := v.Version
	if strings.HasPrefix(strings.ToLower(cleanVersion), "v") {
		cleanVersion = cleanVersion[1:]
	}

	ver, err := version.NewVersion(cleanVersion)
	if err != nil {
		return false, err
	}

	c, err := version.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	return c.Check(ver), nil
}
