package main

import (
	"encoding/base64"
	"io/fs"
	"os"
	"testing"

	"ariga.io/atlas-sync-action/internal/atlascloud"
	"ariga.io/atlas/sql/migrate"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	arc, err := Archive("internal/testdata/basic/migrations")
	require.NoError(t, err)
	exp, err := os.ReadFile("internal/testdata/basic/atlas_archive.tar.b64")
	require.NoError(t, err)
	require.EqualValues(t, string(exp), arc)

	// Test backwards compatability.
	bc, err := os.ReadFile("internal/testdata/basic/bc.tar.b64")
	require.NoError(t, err)
	dec, err := base64.StdEncoding.DecodeString(string(bc))
	require.NoError(t, err)
	dir, err := migrate.UnarchiveDir(dec)
	require.NoError(t, err)
	for _, name := range []string{migrate.HashFileName, "20230201094614.sql"} {
		ex, err := os.ReadFile("internal/testdata/basic/migrations/" + name)
		require.NoError(t, err)
		ac, err := fs.ReadFile(dir, name)
		require.NoError(t, err)
		require.Equal(t, ex, ac)
	}
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
	require.EqualValues(t, atlascloud.ReportDirInput{
		Repo:          "ariga/test",
		Commit:        "1234567890",
		Branch:        "master",
		Path:          "migrations/",
		Driver:        atlascloud.DriverMysql,
		Url:           "https://github.com/ariga/atlas-sync-action/commit/4a3f0bcb6dff19078393728f1b69d89d853771eb",
		DirFormat:     atlascloud.DirFormatAtlas,
		ArchiveFormat: atlascloud.ArchiveFormatB64Tar,
	}, input)
}
