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

type TestRepo struct {
	repo *git.Repository
	test *testing.T
}

func NewTestRepo(t *testing.T, bare bool) TestRepo {
	path, err := ioutil.TempDir("", "gtm")
	CheckFatal(t, err)
	repo, err := git.InitRepository(path, bare)
	CheckFatal(t, err)
	return TestRepo{repo: repo, test: t}
}

func (t TestRepo) Seed() {
	t.SaveFile("README", "", "foo\n")
	treeOid := t.Stage("README")
	t.Commit(treeOid)
	return
}

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
	CheckFatal(t.test, err)
	t.repo.Free()

	return
}

func (t TestRepo) PathIn(name string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(filepath.Dir(t.repo.Path())), name))
}

func (t TestRepo) Stage(files ...string) *git.Oid {
	idx, err := t.repo.Index()
	CheckFatal(t.test, err)
	for _, f := range files {
		err = idx.AddByPath(filepath.ToSlash(f))
		CheckFatal(t.test, err)
	}
	treeId, err := idx.WriteTreeTo(t.repo)
	CheckFatal(t.test, err)
	err = idx.Write()
	CheckFatal(t.test, err)
	return treeId
}

func (t TestRepo) Commit(treeId *git.Oid) *git.Oid {
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
	tree, err := t.repo.LookupTree(treeId)
	CheckFatal(t.test, err)

	var commitId *git.Oid
	if headUnborn {
		commitId, err = t.repo.CreateCommit("HEAD", sig, sig, message, tree)
	} else {
		commitId, err = t.repo.CreateCommit("HEAD", sig, sig, message, tree,
			currentTip)
	}
	CheckFatal(t.test, err)

	return commitId
}

func (t TestRepo) SaveFile(filename, subdir, content string) {
	d := filepath.Join(t.PathIn(""), subdir)
	err := os.MkdirAll(d, 0700)
	CheckFatal(t.test, err)
	err = ioutil.WriteFile(filepath.Join(d, filename), []byte(content), 0644)
	CheckFatal(t.test, err)
}

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
