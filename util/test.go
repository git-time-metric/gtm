// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/libgit2/git2go"
)

// TestRepo represents a test git repo used in testing
type TestRepo struct {
	repo *git.Repository
	test *testing.T
}

// Repo return a pointer to the git repository
func (t TestRepo) Repo() *git.Repository {
	return t.repo
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
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	t.repo.Free()
}

// Workdir return the working directory for the git repository
func (t TestRepo) Workdir() string {
	return filepath.Clean(t.repo.Workdir())
}

// Path return the git path for the git repository
func (t TestRepo) Path() string {
	return filepath.Clean(t.repo.Path())
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
	d := filepath.Join(t.Workdir(), subdir)
	err := os.MkdirAll(d, 0700)
	CheckFatal(t.test, err)
	err = ioutil.WriteFile(filepath.Join(d, filename), []byte(content), 0644)
	CheckFatal(t.test, err)
}

// FileExists Checks if a file exists in the repo folder
func (t TestRepo) FileExists(filename, subdir string) bool {
	_, err := os.Stat(filepath.Join(subdir, filename))
	return !os.IsNotExist(err)
}

// Clone creates a clone of this repo
func (t TestRepo) Clone() TestRepo {
	path, err := ioutil.TempDir("", "gtm")
	CheckFatal(t.test, err)

	r, err := git.Clone(t.repo.Path(), path, &git.CloneOptions{})
	CheckFatal(t.test, err)

	return TestRepo{repo: r, test: t.test}
}

// AddSubmodule adds a submodule to the test repository
func (t TestRepo) AddSubmodule(url, path string) {
	_, err := t.repo.Submodules.Add(url, path, true)
	CheckFatal(t.test, err)
}

func (t TestRepo) remote(name string) *git.Remote {
	remote, err := t.repo.Remotes.Lookup(name)
	CheckFatal(t.test, err)
	return remote
}

// Push to remote refs to remote
func (t TestRepo) Push(name string, refs ...string) {
	if len(refs) == 0 {
		refs = []string{"refs/heads/master"}
	}
	err := t.remote(name).Push(refs, nil)
	CheckFatal(t.test, err)
}

// Fetch refs from remote
func (t TestRepo) Fetch(name string, refs ...string) {
	if len(refs) == 0 {
		refs = []string{"refs/heads/master"}
	}
	err := t.remote(name).Fetch(refs, nil, "")
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
