// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
)

// VerifyCmd contains CLI commands for verify
type VerifyCmd struct {
	Version string
}

// NewVerify returns new VerifyCmd struct with version set
func NewVerify(v string) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return VerifyCmd{Version: v}, nil
	}
}

// Help returns CLI help for Verify command
func (v VerifyCmd) Help() string {
	return v.Synopsis()
}

// Run exectures verify commands with args
func (v VerifyCmd) Run(args []string) int {
	if len(args) == 0 {
		fmt.Println("Unable to verify version, version constraint not provided")
		return 1
	}

	valid, err := v.check(args[0])
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("%t", valid)
	return 0
}

// Synopsis returns verify help
func (v VerifyCmd) Synopsis() string {
	return `
	Usage: gtm verify <version constraint>
	Verify gtm satisfies the version constraint
	This is typically invoked by plug-ins to determine if GTM needs to be upgraded
	`
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
