package main

import (
	"testing"

	"ariga.io/atlas-sync-action/internal/atlascloud"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	arc, err := Archive("internal/testdata/basic/migrations")
	require.NoError(t, err)
	exp := `-- 20230201094614.sql --
create table users (
  id int primary key,
  name varchar(255),
  about text
);
-- atlas.sum --
h1:X97sBPjOeiRWeoEdqpIpHAdzlshqOqllEKrJS9JruPo=
20230201094614.sql h1:GkisBcewe1zpcP5IGGkbbsYSKJgV+fzlDVal+FKEjoE=
`
	require.Equal(t, exp, arc)
}

func TestInput(t *testing.T) {
	env := map[string]string{
		"GITHUB_REPOSITORY": "ariga/test",
		"GITHUB_SHA":        "1234567890",
		"INPUT_DIR":         "migrations/",
		"INPUT_DRIVER":      "mysql",
		"GITHUB_REF_NAME":   "master",
		"GITHUB_EVENT_PATH": "internal/testdata/push_event.json",
	}
	act := githubactions.New(githubactions.WithGetenv(func(key string) string {
		return env[key]
	}))
	input, err := Input(act)
	require.NoError(t, err)
	require.EqualValues(t, atlascloud.UploadDirInput{
		Repo:      "ariga/test",
		Commit:    "1234567890",
		Branch:    "master",
		Path:      "migrations/",
		Driver:    atlascloud.DriverMysql,
		Url:       "https://github.com/ariga/atlas-sync-action/commit/4a3f0bcb6dff19078393728f1b69d89d853771eb",
		DirFormat: atlascloud.DirFormatAtlas,
	}, input)
}
