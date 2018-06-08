package project

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/git-time-metric/gtm/epoch"
	"github.com/git-time-metric/gtm/util"
)

func TestActiveProject(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	defer os.Remove(d)

	if err := SetActive(d); err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	p, err := GetActive()
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	if p != d {
		t.Errorf("want project directory '%s' got '%s'", d, p)
	}

}

func TestInactiveProject(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	defer os.Remove(d)

	if err := SetActive(d); err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	p, err := GetActive()
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	if p != d {
		t.Errorf("want project directory '%s' got '%s'", d, p)
	}

	saveNow := util.Now
	defer func() { util.Now = saveNow }()
	util.Now = func() time.Time { return time.Now().Add(time.Duration(epoch.IdleProjectTimeout) * time.Second) }

	p, err = GetActive()
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	if p != "" {
		t.Errorf("want project directory '' got '%s'", p)
	}

}

func TestInvalidPath(t *testing.T) {
	if err := SetActive("/path/does/not/exist"); err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	p, err := GetActive()
	if err != nil {
		t.Errorf("want error nil got error %s", err)
	}
	if p != "" {
		t.Errorf("want project directory '' got '%s'", p)
	}
}
