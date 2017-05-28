// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package scm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/git-time-metric/gtm/util"
	"github.com/libgit2/git2go"
)

// RootPath discovers the base directory for a git repo
func RootPath(path ...string) (string, error) {
	defer util.TimeTrack(time.Now(), "scm.RootPath")
	var (
		wd  string
		p   string
		err error
	)
	if len(path) > 0 {
		wd = path[0]
	} else {
		wd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	//TODO: benchmark the call to git.Discover
	//TODO: optionally print result with -debug flag
	p, err = git.Discover(wd, false, []string{})
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Dir(filepath.Dir(p))), nil
}

// CommitLimiter struct filter commits by criteria
type CommitLimiter struct {
	Max        int
	DateRange  util.DateRange
	Before     time.Time
	After      time.Time
	Author     string
	Message    string
	HasMax     bool
	HasBefore  bool
	HasAfter   bool
	HasAuthor  bool
	HasMessage bool
}

// NewCommitLimiter returns a new initialize CommitLimiter struct
func NewCommitLimiter(
	max int, fromDateStr, toDateStr, author, message string,
	today, yesterday, thisWeek, lastWeek,
	thisMonth, lastMonth, thisYear, lastYear bool) (CommitLimiter, error) {

	const dateFormat = "2006-01-02"

	fromDateStr = strings.TrimSpace(fromDateStr)
	toDateStr = strings.TrimSpace(toDateStr)
	author = strings.TrimSpace(author)
	message = strings.TrimSpace(message)

	cnt := func(vals []bool) int {
		var c int
		for _, v := range vals {
			if v {
				c++
			}
		}
		return c
	}([]bool{
		fromDateStr != "" || toDateStr != "",
		today, yesterday, thisWeek, lastWeek, thisMonth, lastMonth, thisYear, lastYear})

	if cnt > 1 {
		return CommitLimiter{}, fmt.Errorf("Using multiple temporal flags is not allowed")
	}

	var (
		err       error
		dateRange util.DateRange
	)

	switch {
	case fromDateStr != "" || toDateStr != "":
		fromDate := time.Time{}
		toDate := time.Time{}

		if fromDateStr != "" {
			fromDate, err = time.Parse(dateFormat, fromDateStr)
			if err != nil {
				return CommitLimiter{}, err
			}
		}
		if toDateStr != "" {
			toDate, err = time.Parse(dateFormat, toDateStr)
			if err != nil {
				return CommitLimiter{}, err
			}
		}
		dateRange = util.DateRange{Start: fromDate, End: toDate}

	case today:
		dateRange = util.TodayRange()
	case yesterday:
		dateRange = util.YesterdayRange()
	case thisWeek:
		dateRange = util.ThisWeekRange()
	case lastWeek:
		dateRange = util.LastWeekRange()
	case thisMonth:
		dateRange = util.ThisMonthRange()
	case lastMonth:
		dateRange = util.LastMonthRange()
	case thisYear:
		dateRange = util.ThisYearRange()
	case lastYear:
		dateRange = util.LastYearRange()
	}

	hasMax := max > 0
	hasAuthor := author != ""
	hasMessage := message != ""

	if !(hasMax || dateRange.IsSet() || hasAuthor || hasMessage) {
		// if no limits set default to max of one result
		hasMax = true
		max = 1
	}

	return CommitLimiter{
		DateRange:  dateRange,
		Max:        max,
		Author:     author,
		Message:    message,
		HasMax:     hasMax,
		HasAuthor:  hasAuthor,
		HasMessage: hasMessage,
	}, nil
}

func (m CommitLimiter) filter(c *git.Commit, cnt int) (bool, bool, error) {
	if m.HasMax && m.Max == cnt {
		return false, true, nil
	}

	if m.DateRange.IsSet() && !m.DateRange.Within(c.Author().When) {
		return false, false, nil
	}

	if m.HasAuthor && !strings.Contains(c.Author().Name, m.Author) {
		return false, false, nil
	}

	if m.HasMessage && !(strings.Contains(c.Summary(), m.Message) || strings.Contains(c.Message(), m.Message)) {
		return false, false, nil
	}

	return true, false, nil
}

// CommitIDs returns commit SHA1 IDs starting from the head up to the limit
func CommitIDs(limiter CommitLimiter, wd ...string) ([]string, error) {
	var (
		repo *git.Repository
		cnt  int
		w    *git.RevWalk
		err  error
	)
	commits := []string{}

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}

	if err != nil {
		return commits, err
	}
	defer repo.Free()

	w, err = repo.Walk()
	if err != nil {
		return commits, err
	}
	defer w.Free()

	err = w.PushHead()
	if err != nil {
		return commits, err
	}

	var filterError error

	err = w.Iterate(
		func(commit *git.Commit) bool {
			include, done, err := limiter.filter(commit, cnt)
			if err != nil {
				filterError = err
				return false
			}
			if done {
				return false
			}
			if include {
				commits = append(commits, commit.Object.Id().String())
				cnt++
			}
			return true
		})

	if filterError != nil {
		return commits, filterError
	}
	if err != nil {
		return commits, err
	}

	return commits, nil
}

// Commit contains commit details
type Commit struct {
	ID      string
	OID     *git.Oid
	Summary string
	Message string
	Author  string
	Email   string
	When    time.Time
	Files   []string
	Stats   CommitStats
}

// CommitStats contains the files changed and their stats
type CommitStats struct {
	Files        []string
	Insertions   int
	Deletions    int
	FilesChanged int
}

// ChangeRatePerHour calculates the rate change per hour
func (c CommitStats) ChangeRatePerHour(seconds int) float64 {
	if seconds == 0 {
		return 0
	}
	return (float64(c.Insertions+c.Deletions) / float64(seconds)) * 3600
}

// DiffParentCommit compares commit to it's parent and returns their stats
func DiffParentCommit(childCommit *git.Commit) (CommitStats, error) {
	childTree, err := childCommit.Tree()
	if err != nil {
		return CommitStats{}, err
	}
	defer childTree.Free()

	if childCommit.ParentCount() == 0 {
		// there is no parent commit, should be the first commit in the repo?

		path := ""
		fileCnt := 0
		files := []string{}

		err := childTree.Walk(
			func(s string, entry *git.TreeEntry) int {
				switch entry.Filemode {
				case git.FilemodeTree:
					// directory where file entry is located
					path = filepath.ToSlash(entry.Name)
				default:
					files = append(files, filepath.Join(path, entry.Name))
					fileCnt++
				}
				return 0
			})

		if err != nil {
			return CommitStats{}, err
		}

		return CommitStats{
			Insertions:   fileCnt,
			Deletions:    0,
			Files:        files,
			FilesChanged: fileCnt,
		}, nil
	}

	parentTree, err := childCommit.Parent(0).Tree()
	if err != nil {
		return CommitStats{}, err
	}
	defer parentTree.Free()

	options, err := git.DefaultDiffOptions()
	if err != nil {
		return CommitStats{}, err
	}

	diff, err := childCommit.Owner().DiffTreeToTree(parentTree, childTree, &options)
	if err != nil {
		return CommitStats{}, err
	}
	defer diff.Free()

	files := []string{}
	err = diff.ForEach(
		func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
			// these should only be files that have changed

			files = append(files, filepath.ToSlash(delta.NewFile.Path))

			return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
				return func(line git.DiffLine) error {
					return nil
				}, nil
			}, nil
		}, git.DiffDetailFiles)

	if err != nil {
		return CommitStats{}, err
	}

	stats, err := diff.Stats()
	if err != nil {
		return CommitStats{}, err
	}
	defer stats.Free()

	return CommitStats{
		Insertions:   stats.Insertions(),
		Deletions:    stats.Deletions(),
		Files:        files,
		FilesChanged: stats.FilesChanged(),
	}, err
}

// HeadCommit returns the latest commit
func HeadCommit(wd ...string) (Commit, error) {
	var (
		repo *git.Repository
		err  error
	)
	commit := Commit{}

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}
	if err != nil {
		return commit, err
	}
	defer repo.Free()

	headCommit, err := lookupHeadCommit(repo)
	if err != nil {
		if err == ErrHeadUnborn {
			return commit, nil
		}
		return commit, err
	}
	defer headCommit.Free()

	commitStats, err := DiffParentCommit(headCommit)
	if err != nil {
		return commit, err
	}

	return Commit{
		ID:      headCommit.Object.Id().String(),
		OID:     headCommit.Object.Id(),
		Summary: headCommit.Summary(),
		Message: headCommit.Message(),
		Author:  headCommit.Author().Name,
		Email:   headCommit.Author().Email,
		When:    headCommit.Author().When,
		Stats:   commitStats,
	}, nil
}

// CreateNote creates a git note associated with the head commit
func CreateNote(noteTxt string, nameSpace string, wd ...string) error {
	util.TimeTrack(time.Now(), "scm.CreateNote")

	var (
		repo *git.Repository
		err  error
	)

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}
	if err != nil {
		return err
	}
	defer repo.Free()

	headCommit, err := lookupHeadCommit(repo)
	if err != nil {
		return err
	}

	sig := &git.Signature{
		Name:  headCommit.Author().Name,
		Email: headCommit.Author().Email,
		When:  headCommit.Author().When,
	}

	_, err = repo.Notes.Create("refs/notes/"+nameSpace, sig, sig, headCommit.Id(), noteTxt, false)
	if err != nil {
		return err
	}

	return nil
}

// CommitNote contains a git note's details
type CommitNote struct {
	ID      string
	OID     *git.Oid
	Summary string
	Message string
	Author  string
	Email   string
	When    time.Time
	Note    string
	Stats   CommitStats
}

// ReadNote returns a commit note for the SHA1 commit id
func ReadNote(commitID string, nameSpace string, wd ...string) (CommitNote, error) {
	var (
		err    error
		repo   *git.Repository
		commit *git.Commit
		n      *git.Note
	)

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}

	if err != nil {
		return CommitNote{}, err
	}

	defer func() {
		if commit != nil {
			commit.Free()
		}
		if n != nil {
			n.Free()
		}
		repo.Free()
	}()

	id, err := git.NewOid(commitID)
	if err != nil {
		return CommitNote{}, err
	}

	commit, err = repo.LookupCommit(id)
	if err != nil {
		return CommitNote{}, err
	}

	var noteTxt string
	n, err = repo.Notes.Read("refs/notes/"+nameSpace, id)
	if err != nil {
		noteTxt = ""
	} else {
		noteTxt = n.Message()
	}

	// TODO: should we make this optional for performance reasons?
	stats, err := DiffParentCommit(commit)
	if err != nil {
		return CommitNote{}, err
	}

	return CommitNote{
		ID:      commit.Object.Id().String(),
		OID:     commit.Object.Id(),
		Summary: commit.Summary(),
		Message: commit.Message(),
		Author:  commit.Author().Name,
		Email:   commit.Author().Email,
		When:    commit.Author().When,
		Note:    noteTxt,
		Stats:   stats,
	}, nil
}

// ConfigSet persists git configuration settings
func ConfigSet(settings map[string]string, wd ...string) error {
	var (
		err  error
		repo *git.Repository
		cfg  *git.Config
	)

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}

	cfg, err = repo.Config()
	defer cfg.Free()

	for k, v := range settings {
		err = cfg.SetString(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConfigRemove removes git configuration settings
func ConfigRemove(settings map[string]string, wd ...string) error {
	var (
		err  error
		repo *git.Repository
		cfg  *git.Config
	)

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}

	cfg, err = repo.Config()
	defer cfg.Free()

	for k := range settings {
		err = cfg.Delete(k)
		if err != nil {
			return err
		}
	}
	return nil
}

// GitHook is the Command with options to be added/removed from a git hook
// Exe is the executable file name for Linux/MacOS
// RE is the regex to match on for the command
type GitHook struct {
	Exe     string
	Command string
	RE      *regexp.Regexp
}

func (g GitHook) getCommandPath() string {
	// save current dir &  change to root
	// to guarantee we get the full path
	wd, err := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(string(filepath.Separator))

	p, err := exec.LookPath(g.getExeForOS())
	if err != nil {
		return g.Command
	}
	if runtime.GOOS == "windows" {
		// put "" around file path
		return strings.Replace(g.Command, g.Exe, fmt.Sprintf("%s || \"%s\"", g.Command, p), 1)
	}
	return strings.Replace(g.Command, g.Exe, fmt.Sprintf("%s || %s", g.Command, p), 1)
}

func (g GitHook) getExeForOS() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("gtm.%s", "exe")
	}
	return g.Exe
}

// SetHooks creates git hooks
func SetHooks(hooks map[string]GitHook, wd ...string) error {
	const shebang = "#!/bin/sh"
	for ghfile, hook := range hooks {
		var (
			p   string
			err error
		)

		if len(wd) > 0 {
			p = wd[0]
		} else {
			p, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		fp := filepath.Join(p, ".git", "hooks", ghfile)
		hooksDir := filepath.Join(p, ".git", "hooks")

		var output string

		if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
			if err := os.MkdirAll(hooksDir, 0700); err != nil {
				return err
			}
		}

		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			b, err := ioutil.ReadFile(fp)
			if err != nil {
				return err
			}
			output = string(b)
		}

		if !strings.Contains(output, shebang) {
			output = fmt.Sprintf("%s\n%s", shebang, output)
		}

		if hook.RE.MatchString(output) {
			output = hook.RE.ReplaceAllString(output, fmt.Sprintf("%s", hook.getCommandPath()))
		} else {
			output = fmt.Sprintf("%s\n%s", output, hook.getCommandPath())
		}

		if err = ioutil.WriteFile(fp, []byte(output), 0755); err != nil {
			return err
		}

		if err := os.Chmod(fp, 0755); err != nil {
			return err
		}
	}

	return nil
}

// RemoveHooks remove matching git hook commands
func RemoveHooks(hooks map[string]GitHook, wd ...string) error {
	for ghfile, hook := range hooks {
		var (
			p   string
			err error
		)

		if len(wd) > 0 {
			p = wd[0]
		} else {
			p, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		fp := filepath.Join(p, ".git", "hooks", ghfile)

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		}

		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output := string(b)

		if hook.RE.MatchString(output) {
			output := hook.RE.ReplaceAllString(output, "")
			i := strings.LastIndexAny(output, "\n")
			if i > -1 {
				output = output[0:i]
			}
			if err = ioutil.WriteFile(fp, []byte(output), 0755); err != nil {
				return err
			}
		}

	}

	return nil
}

// IgnoreSet persists paths/files to ignore for a git repo
func IgnoreSet(ignore string, wd ...string) error {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p = wd[0]
	} else {
		p, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	fp := filepath.Join(p, ".gitignore")

	var output string

	data, err := ioutil.ReadFile(fp)
	if err == nil {
		output = string(data)

		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == ignore {
				return nil
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("can't read %s: %s", fp, err)
	}

	if len(output) > 0 && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	output += ignore + "\n"

	if err = ioutil.WriteFile(fp, []byte(output), 0644); err != nil {
		return fmt.Errorf("can't write %s: %s", fp, err)
	}

	return nil
}

// IgnoreRemove removes paths/files ignored for a git repo
func IgnoreRemove(ignore string, wd ...string) error {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p = wd[0]
	} else {
		p, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	fp := filepath.Join(p, ".gitignore")

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return fmt.Errorf("Unable to remove %s from .gitignore, %s not found", ignore, fp)
	}
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	output := string(b)
	if strings.Contains(output, ignore+"\n") {
		output = strings.Replace(output, ignore+"\n", "", 1)
		if err = ioutil.WriteFile(fp, []byte(output), 0644); err != nil {
			return err
		}
	}
	return nil
}

func openRepository(wd ...string) (*git.Repository, error) {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p, err = RootPath(wd[0])
	} else {
		p, err = RootPath()
	}
	if err != nil {
		return nil, err
	}

	repo, err := git.OpenRepository(p)
	return repo, err
}

var (
	// ErrHeadUnborn is raised when there are no commits yet in the git repo
	ErrHeadUnborn = errors.New("Head commit not found")
)

func lookupHeadCommit(repo *git.Repository) (*git.Commit, error) {

	headUnborn, err := repo.IsHeadUnborn()
	if err != nil {
		return nil, err
	}
	if headUnborn {
		return nil, ErrHeadUnborn
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}
	defer headRef.Free()

	commit, err := repo.LookupCommit(headRef.Target())
	if err != nil {
		return nil, err
	}

	return commit, nil
}

// Status contains the git file statuses
type Status struct {
	Files []fileStatus
}

// NewStatus create a Status struct for a git repo
func NewStatus(wd ...string) (Status, error) {
	util.TimeTrack(time.Now(), "scm.NewStatus")

	var (
		repo *git.Repository
		err  error
	)
	status := Status{}

	if len(wd) > 0 {
		repo, err = openRepository(wd[0])
	} else {
		repo, err = openRepository()
	}
	if err != nil {
		return status, err
	}
	defer repo.Free()

	//TODO: research what status options to set
	opts := &git.StatusOptions{}
	opts.Show = git.StatusShowIndexAndWorkdir
	opts.Flags = git.StatusOptIncludeUntracked | git.StatusOptRenamesHeadToIndex | git.StatusOptSortCaseSensitively
	statusList, err := repo.StatusList(opts)

	if err != nil {
		return status, err
	}
	defer statusList.Free()

	cnt, err := statusList.EntryCount()
	if err != nil {
		return status, err
	}

	for i := 0; i < cnt; i++ {
		entry, err := statusList.ByIndex(i)
		if err != nil {
			return status, err
		}
		status.AddFile(entry)
	}

	return status, nil
}

// AddFile adds a StatusEntry for each file in working and staging directories
func (s *Status) AddFile(e git.StatusEntry) {
	var path string
	if e.Status == git.StatusIndexNew ||
		e.Status == git.StatusIndexModified ||
		e.Status == git.StatusIndexDeleted ||
		e.Status == git.StatusIndexRenamed ||
		e.Status == git.StatusIndexTypeChange {
		path = filepath.ToSlash(e.HeadToIndex.NewFile.Path)
	} else {
		path = filepath.ToSlash(e.IndexToWorkdir.NewFile.Path)
	}
	s.Files = append(s.Files, fileStatus{Path: path, Status: e.Status})
}

// HasStaged returns true if there are any files in staging
func (s *Status) HasStaged() bool {
	for _, f := range s.Files {
		if f.InStaging() {
			return true
		}
	}
	return false
}

// IsModified returns true if the file is modified in either working or staging
func (s *Status) IsModified(path string, staging bool) bool {
	path = filepath.ToSlash(path)
	for _, f := range s.Files {
		if path == f.Path && f.InStaging() == staging {
			return f.IsModified()
		}
	}
	return false
}

// IsTracked returns true if file is tracked by the git repo
func (s *Status) IsTracked(path string) bool {
	path = filepath.ToSlash(path)
	for _, f := range s.Files {
		if path == f.Path {
			return f.IsTracked()
		}
	}
	return false
}

type fileStatus struct {
	Status git.Status
	Path   string
}

// InStaging returns true if the file is in staging
func (f fileStatus) InStaging() bool {
	return f.Status == git.StatusIndexNew ||
		f.Status == git.StatusIndexModified ||
		f.Status == git.StatusIndexDeleted ||
		f.Status == git.StatusIndexRenamed ||
		f.Status == git.StatusIndexTypeChange
}

// InWorking returns true if the file is in working
func (f fileStatus) InWorking() bool {
	return f.Status == git.StatusWtModified ||
		f.Status == git.StatusWtDeleted ||
		f.Status == git.StatusWtRenamed ||
		f.Status == git.StatusWtTypeChange
}

// IsTracked returns true if the file is tracked by git
func (f fileStatus) IsTracked() bool {
	return f.Status != git.StatusIgnored &&
		f.Status != git.StatusWtNew
}

// IsModified returns true if the file has been modified
func (f fileStatus) IsModified() bool {
	return f.InStaging() || f.InWorking()
}
