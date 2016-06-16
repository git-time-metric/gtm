package scm

import (
	"os"
	"testing"

	"edgeg.io/gtm/util"
)

func TestRootPath(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()

	repoPath := repo.PathIn("")
	wantPath := repoPath
	gotPath, err := RootPath(repoPath)
	if err != nil {
		t.Errorf("RootPath error, %s", err)
	}
	if wantPath != gotPath {
		t.Errorf("RootPath want %s, got %s", wantPath, gotPath)
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	util.CheckFatal(t, err)

	gotPath, err = RootPath()
	if err != nil {
		t.Errorf("RootPath error, %s", err)
	}
	if wantPath != gotPath {
		t.Errorf("RootPath want %s, got %s", wantPath, gotPath)
	}
}

func TestCommitIDs(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()

	repoPath := repo.PathIn("")

	commits, err := CommitIDs(2, repoPath)
	if err != nil {
		t.Errorf("CommitIDs error, %s", err)
	}
	if len(commits) != 1 {
		t.Errorf("CommitIDs want 1 commit, got %d", len(commits))
	}

	saveDir, err := os.Getwd()
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
	util.CheckFatal(t, err)

	commits, err = CommitIDs(2)
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

	repoPath := repo.PathIn("")

	commit, err := HeadCommit(repoPath)
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

	err = os.Chdir(repoPath)
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

	repoPath := repo.PathIn("")

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
	util.CheckFatal(t, err)
	defer os.Chdir(saveDir)

	err = os.Chdir(repoPath)
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

	note, err = ReadNote(commit.ID, "gtm-data")
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

	repoPath := repo.PathIn("")

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
		t.Errorf("len(status.Files) want \"1\" got \"%d\"", len(status.Files))
	}
}
