// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/libgit2/git2go"
)

var TimeTrackEnable = false

//TimeTrack is used for profiling execution time
func TimeTrack(start time.Time, name string) {
	if TimeTrackEnable {
		elapsed := time.Since(start)
		log.Printf("%s took %s", name, elapsed)
	}
}

// TestRepo represents a test git repo used in testing
type TestRepo struct {
	repo *git.Repository
	test *testing.T
}

// NewTestRepo creates a new instance of TestRepo
func NewTestRepo(t *testing.T, bare bool) TestRepo {
	path, err := ioutil.TempDir("", "gtm")
	CheckFatal(t, err)
	repo, err := git.InitRepository(path, bare)
	CheckFatal(t, err)
	return TestRepo{repo: repo, test: t}
}

// Seed creates test data for the git repo
func (t TestRepo) Seed() {
	t.SaveFile("README", "", "foo\n")
	treeOid := t.Stage("README")
	t.Commit(treeOid)
	return
}

// Remove deletes temp directories, files and git repo
func (t TestRepo) Remove() {
	var repoPath string

	if t.repo.IsBare() {
		repoPath = t.repo.Path()
	} else {
		repoPath = t.repo.Workdir()
	}

	// assert it's in a temp dir just in case
	if !strings.Contains(filepath.Clean(repoPath), filepath.Clean(os.TempDir())) {
		CheckFatal(t.test, fmt.Errorf("Unable to remove, repoPath %s is not within %s", repoPath, os.TempDir()))
		return
	}

	err := os.RemoveAll(repoPath)
	if err != nil {
		// this could be just the issue with Windows os.RemoveAll() and privileges, ignore
		fmt.Fprintln(os.Stderr, err)
	}
	t.repo.Free()

	return
}

// PathIn returns full path of file within repo
func (t TestRepo) PathIn(name string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(filepath.Dir(t.repo.Path())), name))
}

// Stage adds files to staging for git repo
func (t TestRepo) Stage(files ...string) *git.Oid {
	idx, err := t.repo.Index()
	CheckFatal(t.test, err)
	for _, f := range files {
		err = idx.AddByPath(filepath.ToSlash(f))
		CheckFatal(t.test, err)
	}
	treeID, err := idx.WriteTreeTo(t.repo)
	CheckFatal(t.test, err)
	err = idx.Write()
	CheckFatal(t.test, err)
	return treeID
}

// Commit commits staged files
func (t TestRepo) Commit(treeID *git.Oid) *git.Oid {
	loc, err := time.LoadLocation("America/Chicago")
	CheckFatal(t.test, err)
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	headUnborn, err := t.repo.IsHeadUnborn()
	CheckFatal(t.test, err)
	var currentTip *git.Commit

	if !headUnborn {
		currentBranch, err := t.repo.Head()
		CheckFatal(t.test, err)
		currentTip, err = t.repo.LookupCommit(currentBranch.Target())
		CheckFatal(t.test, err)
	}

	message := "This is a commit\n"
	tree, err := t.repo.LookupTree(treeID)
	CheckFatal(t.test, err)

	var commitID *git.Oid
	if headUnborn {
		commitID, err = t.repo.CreateCommit("HEAD", sig, sig, message, tree)
	} else {
		commitID, err = t.repo.CreateCommit("HEAD", sig, sig, message, tree,
			currentTip)
	}
	CheckFatal(t.test, err)

	return commitID
}

// SaveFile creates a file within the git repo project
func (t TestRepo) SaveFile(filename, subdir, content string) {
	d := filepath.Join(t.PathIn(""), subdir)
	err := os.MkdirAll(d, 0700)
	CheckFatal(t.test, err)
	err = ioutil.WriteFile(filepath.Join(d, filename), []byte(content), 0644)
	CheckFatal(t.test, err)
}

// CheckFatal raises a fatal error if error is not nil
func CheckFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}
