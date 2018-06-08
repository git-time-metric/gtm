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

	SetActive(d)
	p := GetActive()
	if p != d {
		t.Errorf("want project directory %s got %s", d, p)
	}

	SetActive("/path/does/not/exist")
	p = GetActive()
	if p != "" {
		t.Errorf("want project directory '' got %s", p)
	}

	SetActive(d)
	saveNow := util.Now
	defer func() { util.Now = saveNow }()
	util.Now = func() time.Time { return time.Now().Add(time.Duration(epoch.IdleProjectTimeout) - 10*time.Second) }
	p = GetActive()
	if p != d {
		t.Errorf("want project directory %s got %s", d, p)
	}
	util.Now = func() time.Time { return time.Now().Add(time.Duration(epoch.IdleProjectTimeout) * time.Second) }
	p = GetActive()
	if p != "" {
		t.Errorf("want project directory '' got %s", p)
	}
}
