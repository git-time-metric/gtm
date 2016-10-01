package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/git-time-metric/gtm/scm"
)

type Index struct {
	Projects map[string]time.Time
}

func NewIndex() (Index, error) {
	i := Index{Projects: map[string]time.Time{}}

	err := i.load()
	if err != nil {
		err := i.save()
		if err != nil {
			return i, err
		}
	}

	return i, nil
}

func (i *Index) Get(tags []string, all bool) ([]string, error) {
	switch {
	case all:
		err := i.clean()
		return i.projects(), err
	case len(tags) > 0:
		if err := i.clean(); err != nil {
			return []string{}, err
		}
		projectsWithTags := []string{}
		for _, p := range i.projects() {
			found, err := i.hasTags(p, tags)
			if err != nil {
				return []string{}, nil
			}
			if found {
				projectsWithTags = append(projectsWithTags, p)
			}
		}
		return projectsWithTags, nil
	default:
		curProjPath, err := scm.RootPath()
		if err != nil {
			return []string{}, err
		}
		if _, ok := i.Projects[curProjPath]; !ok {
			i.add(curProjPath)
			if err := i.save(); err != nil {
				return []string{}, err
			}
		}
		return []string{curProjPath}, nil
	}
}

func (i *Index) add(p string) {
	i.Projects[p] = time.Now()
}

func (i *Index) remove(p string) {
	fmt.Println("Removing", p)
	delete(i.Projects, p)
}

func (i *Index) projects() []string {
	var keys []string
	for k := range i.Projects {
		keys = append(keys, k)
	}
	return keys
}

func (i *Index) path() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".git-time-metric", "project.idx"), nil
}

func (i *Index) load() error {
	p, err := i.path()
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	json.Unmarshal(raw, &i.Projects)

	return nil
}

func (i *Index) save() error {
	bytes, err := json.Marshal(i.Projects)
	if err != nil {
		return err
	}

	p, err := i.path()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Dir(p)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(p, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (i *Index) hasTags(projectPath string, tagsToFind []string) (bool, error) {
	tags, err := loadTags(filepath.Join(projectPath, ".gtm"))
	if err != nil {
		return false, err
	}
	for _, t1 := range tags {
		for _, t2 := range tagsToFind {
			if t1 == t2 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (i *Index) removeNotFound(projectPath string) bool {
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		i.remove(projectPath)
		return true
	}
	return false
}

func (i *Index) clean() error {
	for _, p := range i.projects() {
		i.removeNotFound(p)
	}
	err := i.save()
	return err
}
