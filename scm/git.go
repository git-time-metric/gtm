package scm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/libgit2/git2go"
)

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

func CommitIDs(limit int, wd ...string) ([]string, error) {
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

	err = w.Iterate(
		func(commit *git.Commit) bool {
			if limit == cnt {
				return false
			}
			commits = append(commits, commit.Object.Id().String())
			cnt++
			return true
		})

	if err != nil {
		return commits, err
	}

	return commits, nil
}

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

func Config(settings map[string]string, wd ...string) error {
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

func SetHooks(hooks map[string]string, wd ...string) error {
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

		var output string
		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			b, err := ioutil.ReadFile(fp)
			if err != nil {
				return err
			}
			output = string(b)

			if strings.Contains(output, command+"\n") {
				// if file already exists this will make sure it's executable
				if err := os.Chmod(fp, 0755); err != nil {
					return err
				}
				return nil
			}
		}

		if err = ioutil.WriteFile(
			fp, []byte(fmt.Sprintf("%s\n%s\n", output, command)), 0755); err != nil {
			return err
		}
		// if file already exists this will make sure it's executable
		if err := os.Chmod(fp, 0755); err != nil {
			return err
		}
	}

	return nil
}

func Ignore(ignore string, wd ...string) error {
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
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output = string(b)

		if strings.Contains(output, ignore+"\n") {
			return nil
		}
	}

	if err = ioutil.WriteFile(
		fp, []byte(fmt.Sprintf("%s\n%s\n", output, ignore)), 0644); err != nil {
		return err
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

type Status struct {
	Files []FileStatus
}

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

	opts := &git.StatusOptions{Show: git.StatusShowIndexAndWorkdir}
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
	s.Files = append(s.Files, FileStatus{Path: path, Status: e.Status})
}

func (s *Status) HasStaged() bool {
	for _, f := range s.Files {
		if f.InStaging() {
			return true
		}
	}
	return false
}

func (s *Status) IsModified(path string, staging bool) bool {
	path = filepath.ToSlash(path)
	for _, f := range s.Files {
		if path == f.Path && f.InStaging() == staging {
			return f.IsModified()
		}
	}
	return false
}

func (s *Status) IsTracked(path string) bool {
	path = filepath.ToSlash(path)
	for _, f := range s.Files {
		if path == f.Path {
			return f.IsTracked()
		}
	}
	return false
}

type FileStatus struct {
	Status git.Status
	Path   string
}

func (f FileStatus) InStaging() bool {
	return f.Status == git.StatusIndexNew ||
		f.Status == git.StatusIndexModified ||
		f.Status == git.StatusIndexDeleted ||
		f.Status == git.StatusIndexRenamed ||
		f.Status == git.StatusIndexTypeChange
}

func (f FileStatus) InWorking() bool {
	return f.Status == git.StatusWtModified ||
		f.Status == git.StatusWtDeleted ||
		f.Status == git.StatusWtRenamed ||
		f.Status == git.StatusWtTypeChange
}

func (f FileStatus) IsTracked() bool {
	return f.Status != git.StatusIgnored &&
		f.Status != git.StatusWtNew
}

func (f FileStatus) IsModified() bool {
	return f.InStaging() || f.InWorking()
}
