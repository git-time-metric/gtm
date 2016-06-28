package note

import (
	"reflect"
	"testing"
)

func TestUnMarshallTimeLog(t *testing.T) {

	cases := []struct {
		Note string
		Want CommitNote
	}{
		{
			`
[ver:1,total:1425]
environment/drone/run-tests.sh:725,1460066400:705,1460070000:20,m
environment/drone/run-tests-cron.sh:700,1460066400:540,1460070000:160,m
`,
			CommitNote{
				Files: []FileDetail{
					FileDetail{
						SourceFile: "environment/drone/run-tests.sh",
						TimeSpent:  725,
						Timeline:   map[int64]int{int64(1460066400): 705, int64(1460070000): 20},
						Status:     "m"},
					FileDetail{
						SourceFile: "environment/drone/run-tests-cron.sh",
						TimeSpent:  700,
						Timeline:   map[int64]int{int64(1460066400): 540, int64(1460070000): 160},
						Status:     "m"},
				},
			},
		},
		{
			`

[ver:1,total:1425]
environment/drone/run-tests.sh:725,1460066400:705,1460070000:20,m
environment/drone/run-tests-cron.sh:700,1460066400:540,1460070000:160,m

[ver:1,total:60]
environment/drone/test.go:60,1460070000:60,r

`,
			CommitNote{
				Files: []FileDetail{
					FileDetail{
						SourceFile: "environment/drone/run-tests.sh",
						TimeSpent:  725,
						Timeline:   map[int64]int{int64(1460066400): 705, int64(1460070000): 20},
						Status:     "m"},
					FileDetail{
						SourceFile: "environment/drone/run-tests-cron.sh",
						TimeSpent:  700,
						Timeline:   map[int64]int{int64(1460066400): 540, int64(1460070000): 160},
						Status:     "m"},
					FileDetail{
						SourceFile: "environment/drone/test.go",
						TimeSpent:  60,
						Timeline:   map[int64]int{int64(1460070000): 60},
						Status:     "r"},
				},
			},
		},
	}

	for _, tc := range cases {
		got, err := UnMarshal(tc.Note)
		if err != nil {
			t.Errorf("unMarshalTimelog(%s), want error nil got error %s", tc.Note, err)
		}
		if !reflect.DeepEqual(tc.Want, got) {
			t.Errorf("unMarshalTimelog(%s), want:\n%+v\n got:\n%+v\n", tc.Note, tc.Want, got)
		}
	}

}
