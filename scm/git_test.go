package scm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/libgit2/git2go"
)

func TestRootPath(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	repoPath := pathInRepo(repo, "")
	wantPath := repoPath
	gotPath, err := RootPath(repoPath)
	if err != nil {
		t.Errorf("RootPath error, %s", err)
	}
	if wantPath != gotPath {
		t.Errorf("RootPath want %s, got %s", wantPath, gotPath)
	}

	saveDir, err := os.Getwd()
	checkFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	checkFatal(t, err)

	gotPath, err = RootPath()
	if err != nil {
		t.Errorf("RootPath error, %s", err)
	}
	if wantPath != gotPath {
		t.Errorf("RootPath want %s, got %s", wantPath, gotPath)
	}
}

func TestCommitIDs(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	repoPath := pathInRepo(repo, "")

	commits, err := CommitIDs(2, repoPath)
	if err != nil {
		t.Errorf("CommitIDs error, %s", err)
	}
	if len(commits) != 1 {
		t.Errorf("CommitIDs want 1 commit, got %s", len(commits))
	}

	saveDir, err := os.Getwd()
	checkFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	checkFatal(t, err)

	commits, err = CommitIDs(2)
	if err != nil {
		t.Errorf("CommitIDs error, %s", err)
	}
	if len(commits) != 1 {
		t.Errorf("CommitIDs want 1 commit, got %s", len(commits))
	}
}

func TestHeadCommit(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	repoPath := pathInRepo(repo, "")

	commit, err := HeadCommit(repoPath)
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}
	email := "random@hacker.com"
	if commit.Email != email {
		t.Errorf("HeadCommit want email \"%s\", got \"%s\"", email, commit.Email)
	}

	saveDir, err := os.Getwd()
	checkFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	checkFatal(t, err)

	commit, err = HeadCommit()
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}
	if commit.Email != email {
		t.Errorf("HeadCommit want message \"%s\", got \"%s\"", email, commit.Email)
	}
}

func TestNote(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	repoPath := pathInRepo(repo, "")

	noteTxt := "This is a note"
	err := CreateNote(noteTxt, "gtm-data", repoPath)
	if err != nil {
		t.Errorf("CreateNote error, %s", err)
	}

	commit, err := HeadCommit(repoPath)
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}

	note, err := ReadNote(commit.ID, "gtm-data", repoPath)
	if err != nil {
		t.Errorf("ReadNote error, %s", err)
	}

	if note.Note != noteTxt {
		t.Errorf("ReadNote want message \"%s\", got \"%s\"", noteTxt, note.Note)
	}

	saveDir, err := os.Getwd()
	checkFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	checkFatal(t, err)

	err = CreateNote(noteTxt, "gtm-data")
	// Expect error, note should already exist
	if err == nil {
		t.Errorf("CreateNote expected error but got nil")
	}

	commit, err = HeadCommit()
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}

	note, err = ReadNote(commit.ID, "gtm-data")
	if err != nil {
		t.Errorf("ReadNote error, %s", err)
	}

	if note.Note != noteTxt {
		t.Errorf("ReadNote want message \"%s\", got \"%s\"", noteTxt, note.Note)
	}

}

func TestStatus(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)
	updateReadmeInStaging(t, repo, "Test status")

	repoPath := pathInRepo(repo, "")

	status, err := NewStatus(repoPath)
	if err != nil {
		t.Errorf("NewStatus error, %s", err)
	}
	if status.IsModified("README", false) {
		t.Error("status.IsModified want \"false\" got \"true\"")
	}
	if !status.IsModified("README", true) {
		t.Error("status.IsModified want \"true\" got \"false\"")
	}
	if !status.IsTracked("README") {
		t.Error("status.IsTracked want \"true\" got \"false\"")
	}
	if !status.HasStaged() {
		t.Error("status.HasStaged() want \"true\" got \"false\"")
	}
	if len(status.Files) != 1 {
		t.Errorf("len(status.Files) want \"1\" got \"%s\"", len(status.Files))
	}
}

// Test setup/cleanup helper methods copied from git2go
// https://github.com/libgit2/git2go/blob/master/git_test.go

func cleanupTestRepo(t *testing.T, r *git.Repository) {
	var err error
	if r.IsBare() {
		err = os.RemoveAll(r.Path())
	} else {
		err = os.RemoveAll(r.Workdir())
	}
	checkFatal(t, err)

	r.Free()
}

func createTestRepo(t *testing.T) *git.Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "gtm")
	checkFatal(t, err)
	repo, err := git.InitRepository(path, false)
	checkFatal(t, err)

	tmpfile := "README"
	err = ioutil.WriteFile(path+"/"+tmpfile, []byte("foo\n"), 0644)

	checkFatal(t, err)

	return repo
}

func createBareTestRepo(t *testing.T) *git.Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "gtm")
	checkFatal(t, err)
	repo, err := git.InitRepository(path, true)
	checkFatal(t, err)

	return repo
}

func seedTestRepo(t *testing.T, repo *git.Repository) (*git.Oid, *git.Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	return commitId, treeId
}

func pathInRepo(repo *git.Repository, name string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(filepath.Dir(repo.Path())), name))
}

func updateReadmeInStaging(t *testing.T, repo *git.Repository, content string) {
	tmpfile := "README"
	err := ioutil.WriteFile(pathInRepo(repo, tmpfile), []byte(content), 0644)
	checkFatal(t, err)

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	_, err = idx.WriteTree()
	checkFatal(t, err)

	return
}

func updateReadme(t *testing.T, repo *git.Repository, content string) (*git.Oid, *git.Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	tmpfile := "README"
	err = ioutil.WriteFile(pathInRepo(repo, tmpfile), []byte(content), 0644)
	checkFatal(t, err)

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	currentBranch, err := repo.Head()
	checkFatal(t, err)
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	checkFatal(t, err)

	return commitId, treeId
}

func checkFatal(t *testing.T, err error) {
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
