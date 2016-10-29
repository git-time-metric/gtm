package scm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-time-metric/git2go"
	"github.com/jinzhu/now"
)

// RootPath discovers the base directory for a git repo
func RootPath(path ...string) (string, error) {
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
	p, err = git.Discover(wd, false, []string{})
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Dir(filepath.Dir(p))), nil
}

type commitLimiter struct {
	Max        int
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

func NewCommitLimiter(
	max int, beforeStr, afterStr, author, message string,
	today, yesterday, thisWeek, lastWeek,
	thisMonth, lastMonth, thisYear, lastYear bool) (commitLimiter, error) {

	const dateFormat = "2006-01-02"

	beforeStr = strings.TrimSpace(beforeStr)
	afterStr = strings.TrimSpace(afterStr)
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
		beforeStr != "" || afterStr != "",
		today, yesterday, thisWeek, lastWeek, thisMonth, lastMonth, thisYear, lastYear})

	if cnt > 1 {
		return commitLimiter{}, fmt.Errorf("Using multiple temporal flags is not allowed")
	}

	var err error
	after := time.Time{}
	before := time.Time{}

	switch {
	case beforeStr != "" || afterStr != "":
		if beforeStr != "" {
			before, err = time.Parse(dateFormat, beforeStr)
			if err != nil {
				return commitLimiter{}, err
			}
		}
		if afterStr != "" {
			after, err = time.Parse(dateFormat, afterStr)
			if err != nil {
				return commitLimiter{}, err
			}
		}
	case today:
		after = now.EndOfDay().AddDate(0, 0, -1)
		// fmt.Println("after", after)
	case yesterday:
		before = now.BeginningOfDay()
		after = now.EndOfDay().AddDate(0, 0, -2)
		// fmt.Println("before", before, "after", after)
	case thisWeek:
		after = now.EndOfWeek().AddDate(0, 0, -7)
		// fmt.Println("after", after)
	case lastWeek:
		before = now.BeginningOfWeek()
		after = now.EndOfWeek().AddDate(0, 0, -14)
		// fmt.Println("before", before, "after", after)
	case thisMonth:
		after = now.EndOfMonth().AddDate(0, -1, -1)
		// fmt.Println("after", after)
	case lastMonth:
		before = now.BeginningOfMonth()
		after = now.EndOfMonth().AddDate(0, -2, 0)
		// fmt.Println("before", before, "after", after)
	case thisYear:
		after = now.EndOfYear().AddDate(-1, 0, 0)
		// fmt.Println("after", after)
	case lastYear:
		before = now.BeginningOfYear()
		after = now.EndOfYear().AddDate(-2, 0, 0)
		// fmt.Println("before", before, "after", after)
	}

	hasMax := max > 0
	hasBefore := !before.IsZero()
	hasAfter := !after.IsZero()
	hasAuthor := author != ""
	hasMessage := message != ""

	if hasBefore && hasAfter && before.Before(after) {
		return commitLimiter{}, fmt.Errorf("Before %s can not be older than after %s", before, after)
	}

	if !(hasMax || hasBefore || hasAfter || hasAuthor || hasMessage) {
		// if no limits set default to max of one result
		hasMax = true
		max = 1
	}

	return commitLimiter{
		Max:        max,
		Before:     before,
		After:      after,
		Author:     author,
		Message:    message,
		HasMax:     hasMax,
		HasBefore:  hasBefore,
		HasAfter:   hasAfter,
		HasAuthor:  hasAuthor,
		HasMessage: hasMessage,
	}, nil
}

func (l commitLimiter) filter(c *git.Commit, cnt int) (bool, bool, error) {
	if l.HasMax && l.Max == cnt {
		return false, true, nil
	}

	if l.HasBefore && !c.Author().When.Before(l.Before) {
		return false, false, nil
	}

	if l.HasAfter && !c.Author().When.After(l.After) {
		return false, true, nil
	}

	if l.HasAuthor && !strings.Contains(c.Author().Name, l.Author) {
		return false, false, nil
	}

	if l.HasMessage && !(strings.Contains(c.Summary(), l.Message) || strings.Contains(c.Message(), l.Message)) {
		return false, false, nil
	}

	return true, false, nil
}

// CommitIDs returns commit SHA1 IDs starting from the head up to the limit
func CommitIDs(limiter commitLimiter, wd ...string) ([]string, error) {
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

	headTree, err := headCommit.Tree()
	if err != nil {
		return commit, err
	}
	defer headTree.Free()

	files := []string{}
	if headCommit.ParentCount() > 0 {
		parentTree, err := headCommit.Parent(0).Tree()
		if err != nil {
			return commit, err
		}
		defer parentTree.Free()

		options, err := git.DefaultDiffOptions()
		if err != nil {
			return commit, err
		}

		diff, err := headCommit.Owner().DiffTreeToTree(parentTree, headTree, &options)
		if err != nil {
			return commit, err
		}
		defer diff.Free()

		err = diff.ForEach(
			func(file git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {

				files = append(files, filepath.ToSlash(file.NewFile.Path))

				return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
					return func(line git.DiffLine) error {
						return nil
					}, nil
				}, nil
			}, git.DiffDetailFiles)

		if err != nil {
			return commit, err
		}

	} else {

		path := ""
		err := headTree.Walk(
			func(s string, entry *git.TreeEntry) int {
				switch entry.Filemode {
				case git.FilemodeTree:
					path = filepath.ToSlash(entry.Name)
				default:
					files = append(files, filepath.Join(path, entry.Name))
				}
				return 0
			})

		if err != nil {
			return commit, err
		}
	}

	commit = Commit{
		ID:      headCommit.Object.Id().String(),
		OID:     headCommit.Object.Id(),
		Summary: headCommit.Summary(),
		Message: headCommit.Message(),
		Author:  headCommit.Author().Name,
		Email:   headCommit.Author().Email,
		When:    headCommit.Author().When,
		Files:   files}

	return commit, nil
}

// CreateNote creates a git note associated with the head commit
func CreateNote(noteTxt string, nameSpace string, wd ...string) error {
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

	return CommitNote{
		ID:      commit.Object.Id().String(),
		OID:     commit.Object.Id(),
		Summary: commit.Summary(),
		Message: commit.Message(),
		Author:  commit.Author().Name,
		Email:   commit.Author().Email,
		When:    commit.Author().When,
		Note:    noteTxt,
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

// SetHooks creates git hooks
func SetHooks(hooks map[string]string, wd ...string) error {
	const shebang = "#!/bin/sh"
	for hook, command := range hooks {
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
		fp := filepath.Join(p, ".git", "hooks", hook)
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

		if !strings.Contains(output, command) {
			output = fmt.Sprintf("%s\n%s\n", output, command)
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
func RemoveHooks(hooks map[string]string, wd ...string) error {
	for hook, command := range hooks {
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
		fp := filepath.Join(p, ".git", "hooks", hook)

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		}

		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output := string(b)

		if strings.Contains(output, command) {
			output = strings.Replace(output, command, "", -1)
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
