package event

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-time-metric/gtm/project"
	"github.com/git-time-metric/gtm/scm"
	"github.com/git-time-metric/gtm/util"
)

//Clean removes any event or metrics files from project in the current working directory
func Clean(dr util.DateRange, application, editor, terminal bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return fmt.Errorf("Unable to clean, Git repository not found in %s", projRoot)
	}

	gtmPath := filepath.Join(projRoot, project.GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return fmt.Errorf("Unable to clean GTM data, %s directory not found", gtmPath)
	}

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".event") &&
			!strings.HasSuffix(f.Name(), ".metric") {
			continue
		}

		if !dr.Within(f.ModTime()) {
			continue
		}

		p := filepath.Join(gtmPath, f.Name())
		a := NewApplicationFromPath(p)
		if !terminal && a.IsTerminal() {
			continue
		}
		if !application && a.IsApplication() && !a.IsTerminal() {
			continue
		}

		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}
