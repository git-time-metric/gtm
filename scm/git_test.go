// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package scm

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/git-time-metric/gtm/util"
)

func TestWorkdir(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	repo.AddSubmodule("http://example.org/submodule", "submodule")

	gotPath, err := Workdir(repo.Path())
	if err != nil {
		t.Errorf("Workdir error, %s", err)
	}
	if repo.Workdir() != gotPath {
		t.Errorf("Workdir want %s, got %s", repo.Workdir(), gotPath)
	}

	sdir := filepath.Join(repo.Workdir(), "submodule")
	gotPath, err = Workdir(sdir)
	if err != nil {
		t.Errorf("Workdir want error nil got %s", err)
	}
	if sdir != gotPath {
		t.Errorf("Workdir want %s for submodule, got %s", sdir, gotPath)
	}
}

func TestGitRepoPath(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	repo.AddSubmodule("http://example.org/submodule", "submodule")

	gotPath, err := GitRepoPath(repo.Path())
	if err != nil {
		t.Errorf("GitRepoPath error, %s", err)
	}
	if repo.Path() != gotPath {
		t.Errorf("GitRepoPath want %s, got %s", repo.Path(), gotPath)
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	os.Chdir(repo.Path())
	gotPath, err = GitRepoPath()
	if err != nil {
		t.Errorf("GitRepoPath error, %s", err)
	}
	if repo.Path() != gotPath {
		t.Errorf("GitRepoPath want %s, got %s", repo.Path(), gotPath)
	}

	subGitdir := filepath.Join(repo.Workdir(), ".git", "modules", "submodule")
	gotPath, err = GitRepoPath(filepath.Join(repo.Workdir(), "submodule"))
	if err != nil {
		t.Errorf("GitRepoPath want error nil got %s", err)
	}
	if subGitdir != gotPath {
		t.Errorf("GitRepoPath want %s for submodule, got %s", subGitdir, gotPath)
	}
}

func TestCommitIDs(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	workdir := repo.Workdir()

	commits, err := CommitIDs(CommitLimiter{Max: 2}, workdir)
	if err != nil {
		t.Errorf("CommitIDs error, %s", err)
	}
	if len(commits) != 1 {
		t.Errorf("CommitIDs want 1 commit, got %d", len(commits))
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(workdir)
	util.CheckFatal(t, err)

	commits, err = CommitIDs(CommitLimiter{Max: 2})
	if err != nil {
		t.Errorf("CommitIDs error, %s", err)
	}
	if len(commits) != 1 {
		t.Errorf("CommitIDs want 1 commit, got %d", len(commits))
	}
}

func TestHeadCommit(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	workdir := repo.Workdir()

	commit, err := HeadCommit(workdir)
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}
	email := "random@hacker.com"
	if commit.Email != email {
		t.Errorf("HeadCommit want email \"%s\", got \"%s\"", email, commit.Email)
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(workdir)
	util.CheckFatal(t, err)

	commit, err = HeadCommit()
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}
	if commit.Email != email {
		t.Errorf("HeadCommit want message \"%s\", got \"%s\"", email, commit.Email)
	}
}

func TestNote(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	workdir := repo.Workdir()

	noteTxt := "This is a note"
	err := CreateNote(noteTxt, "gtm-data", workdir)
	if err != nil {
		t.Errorf("CreateNote error, %s", err)
	}

	commit, err := HeadCommit(workdir)
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}

	note, err := ReadNote(commit.ID, "gtm-data", true, workdir)
	if err != nil {
		t.Errorf("ReadNote error, %s", err)
	}

	if note.Note != noteTxt {
		t.Errorf("ReadNote want message \"%s\", got \"%s\"", noteTxt, note.Note)
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(workdir)
	util.CheckFatal(t, err)

	err = CreateNote(noteTxt, "gtm-data")
	// Expect error, note should already exist
	if err == nil {
		t.Errorf("CreateNote expected error but got nil")
	}

	commit, err = HeadCommit()
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}

	note, err = ReadNote(commit.ID, "gtm-data", true)
	if err != nil {
		t.Errorf("ReadNote error, %s", err)
	}

	if note.Note != noteTxt {
		t.Errorf("ReadNote want message \"%s\", got \"%s\"", noteTxt, note.Note)
	}

}

func TestStatus(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repo.SaveFile("README", "", "Updated readme file")
	repo.Stage("README")

	workdir := repo.Workdir()

	status, err := NewStatus(workdir)
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
		t.Errorf("len(status.Files) want \"1\" got \"%d\"", len(status.Files))
	}
}

func TestIgnoreSet_GitignoreDoesNotExists(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	workdir := repo.Workdir()
	gitignorePath := filepath.Join(workdir, ".gitignore")

	err := IgnoreSet("/.gtm/", workdir)
	if err != nil {
		t.Errorf("IgnoreSet error: %s", err)
	}

	data, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		t.Errorf("read .gitignore error: %s", err)
	}

	if string(data) != "/.gtm/\n" {
		t.Errorf(
			".gitignore want contents \"/.gtm/\n\", got \"%s\"",
			string(data),
		)
	}
}

func TestIgnoreSet_GitignoreIsEmpty(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	workdir := repo.Workdir()
	gitignorePath := filepath.Join(workdir, ".gitignore")

	_, err := os.Create(gitignorePath)
	if err != nil {
		t.Errorf("can't create .gitignore: %s", err)
	}

	err = IgnoreSet("/.gtm/", workdir)
	if err != nil {
		t.Errorf("IgnoreSet error: %s", err)
	}

	data, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		t.Errorf("read .gitignore error: %s", err)
	}

	if string(data) != "/.gtm/\n" {
		t.Errorf(
			".gitignore want contents \"/.gtm/\n\", got \"%s\"",
			string(data),
		)
	}
}

func TestIgnoreSet_GitignoreContainsSomeData(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	workdir := repo.Workdir()
	gitignorePath := filepath.Join(workdir, ".gitignore")

	err := ioutil.WriteFile(gitignorePath, []byte("blah\n"), 0644)
	if err != nil {
		t.Errorf("can't create .gitignore: %s", err)
	}

	err = IgnoreSet("/.gtm/", workdir)
	if err != nil {
		t.Errorf("IgnoreSet error: %s", err)
	}

	data, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		t.Errorf("read .gitignore error: %s", err)
	}

	if string(data) != "blah\n/.gtm/\n" {
		t.Errorf(
			".gitignore want contents \"blah\n/.gtm/\n\", got \"%s\"",
			string(data),
		)
	}
}

func TestIgnoreSet_GitignoreAlreadyContainsGivenData(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	workdir := repo.Workdir()
	gitignorePath := filepath.Join(workdir, ".gitignore")

	err := ioutil.WriteFile(gitignorePath, []byte("/.gtm/\n"), 0644)
	if err != nil {
		t.Errorf("can't create .gitignore: %s", err)
	}

	err = IgnoreSet("/.gtm/", workdir)
	if err != nil {
		t.Errorf("IgnoreSet error: %s", err)
	}

	data, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		t.Errorf("read .gitignore error: %s", err)
	}

	if string(data) != "/.gtm/\n" {
		t.Errorf(
			".gitignore want contents \"/.gtm/\n\", got \"%s\"",
			string(data),
		)
	}
}

func TestIgnoreSet_GitignoreError(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	workdir := repo.Workdir()
	gitignorePath := filepath.Join(workdir, ".gitignore")

	// create directory with name .gitignore for io read error
	err := os.Mkdir(gitignorePath, 0644)
	if err != nil {
		t.Errorf("can't create directory %s: %s", gitignorePath, err)
	}

	err = IgnoreSet("/.gtm/", workdir)
	if err == nil {
		t.Errorf("IgnoreSet must return error, .gitignore is error")
	}

	if !strings.Contains(err.Error(), "can't read") {
		t.Errorf(
			"IgnoreSet error must contain \"can't read\", got \"%s\"",
			err,
		)
	}
}

func TestSetGitHooks(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	gitRepoPath := repo.GitRepoPath()

	hooks := map[string]GitHook{
		"post-commit": {
			Exe:     "gtm",
			Command: "gtm commit --yes",
			RE:      regexp.MustCompile(`(?s)[/,:,a-z,A-Z,0-9,$,-,_,=, ]*gtm\s+commit\s+--yes\.*`)},
	}

	// test when hook exists
	err := ioutil.WriteFile(path.Join(gitRepoPath, "hooks", "post-commit"), []byte{}, 0755)
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}

	err = SetHooks(hooks, gitRepoPath)
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}
	b, err := ioutil.ReadFile(path.Join(gitRepoPath, "hooks", "post-commit"))
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}
	output := string(b)
	if !strings.Contains(output, hooks["post-commit"].Command) {
		t.Errorf("SetHooks(hooks) expected post-commit to contain %s, got %s", hooks["post-commit"].Command, output)
	}

	// test if hook doesn't exist
	err = os.Remove(path.Join(gitRepoPath, "hooks", "post-commit"))
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}

	err = SetHooks(hooks, gitRepoPath)
	if err != nil {
		t.Errorf("SetHooks(hooks) expect error nil, got %s", err)
	}
	b, err = ioutil.ReadFile(path.Join(gitRepoPath, "hooks", "post-commit"))
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}
	output = string(b)
	if !strings.Contains(output, hooks["post-commit"].Command) {
		t.Errorf("SetHooks(hooks) expected post-commit to contain %s, got %s", hooks["post-commit"].Command, output)
	}

	// test if hooks folder doesn't exist
	err = os.RemoveAll(path.Join(gitRepoPath, "hooks"))
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}

	err = SetHooks(hooks, gitRepoPath)
	if err != nil {
		t.Errorf("SetHooks(hooks) expect error nil, got %s", err)
	}
	b, err = ioutil.ReadFile(path.Join(gitRepoPath, "hooks", "post-commit"))
	if err != nil {
		t.Fatalf("SetHooks(hooks) expect error nil, got %s", err)
	}
	output = string(b)
	if !strings.Contains(output, hooks["post-commit"].Command) {
		t.Errorf("SetHooks(hooks) expected post-commit to contain %s, got %s", hooks["post-commit"].Command, output)
	}

}

func TestPushFetchRemote(t *testing.T) {
	remoteRepo := util.NewTestRepo(t, true)
	defer remoteRepo.Remove()

	localRepo := remoteRepo.Clone()
	defer localRepo.Remove()
	localRepo.Seed()

	noteTxt := "This is a note"
	err := CreateNote(noteTxt, "gtm-data", localRepo.Workdir())
	if err != nil {
		t.Errorf("CreateNote error, %s", err)
	}

	localRepo.Push("origin", "refs/heads/master", "refs/notes/gtm-data")

	// clone remote again in another directory
	localRepo2 := remoteRepo.Clone()
	defer localRepo2.Remove()
	localRepo2.Fetch("origin", "refs/notes/gtm-data:refs/notes/gtm-data")

	commit, err := HeadCommit(localRepo2.Workdir())
	if err != nil {
		t.Errorf("HeadCommit error, %s", err)
	}

	note, err := ReadNote(commit.ID, "gtm-data", true, localRepo2.Workdir())
	if err != nil {
		t.Errorf("ReadNote error, %s", err)
	}

	if note.Note != noteTxt {
		t.Errorf("ReadNote want message \"%s\", got \"%s\"", noteTxt, note.Note)
	}
}
