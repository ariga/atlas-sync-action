package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"ariga.io/atlas-sync-action/internal/atlascloud"
	"ariga.io/atlas/sql/migrate"
	"github.com/mitchellh/mapstructure"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/tools/txtar"
)

func main() {
	act := githubactions.New()
	input, err := Input(act)
	if err != nil {
		act.Fatalf("failed to parse input: %v", err)
	}
	githubactions.Noticef("%v", input)
}

// Archive returns a txtar archive of the given migration directory.
func Archive(path string) (string, error) {
	d, err := migrate.NewLocalDir(path)
	if err != nil {
		return "", err
	}
	files, err := d.Files()
	if err != nil {
		return "", err
	}
	arc := &txtar.Archive{}
	for _, f := range files {
		arc.Files = append(arc.Files, txtar.File{
			Name: f.Name(),
			Data: f.Bytes(),
		})
	}
	sumf, err := d.Open(migrate.HashFileName)
	if err != nil {
		return "", fmt.Errorf("opening sumfile: %w", err)
	}
	curS, err := io.ReadAll(sumf)
	if err != nil {
		return "", fmt.Errorf("reading sumfile: %w", err)
	}
	sum, err := d.Checksum()
	if err != nil {
		return "", err
	}
	wantS, err := sum.MarshalText()
	if err != nil {
		return "", err
	}
	if !bytes.Equal(curS, wantS) {
		return "", migrate.ErrChecksumMismatch
	}
	arc.Files = append(arc.Files, txtar.File{
		Name: migrate.HashFileName,
		Data: wantS,
	})
	return string(txtar.Format(arc)), nil
}

func Input(act *githubactions.Action) (atlascloud.UploadDirInput, error) {
	c, err := act.Context()
	if err != nil {
		return atlascloud.UploadDirInput{}, err
	}
	org, repo := c.Repo()
	ev := PushEvent{}
	if err := mapstructure.Decode(c.Event, &ev); err != nil {
		return atlascloud.UploadDirInput{}, err
	}
	di := act.GetInput("driver")
	drv, err := driver(di)
	if err != nil {
		return atlascloud.UploadDirInput{}, err
	}
	return atlascloud.UploadDirInput{
		Repo:      fmt.Sprintf("%s/%s", org, repo),
		Branch:    c.RefName,
		Commit:    c.SHA,
		Path:      act.GetInput("dir"),
		Url:       ev.HeadCommit.URL,
		Driver:    drv,
		DirFormat: atlascloud.DirFormatAtlas,
	}, nil
}

func driver(s string) (atlascloud.Driver, error) {
	switch s := strings.ToLower(s); s {
	case "postgres":
		return atlascloud.DriverPostgresql, nil
	case "mysql":
		return atlascloud.DriverMysql, nil
	case "sqlite":
		return atlascloud.DriverSqlite, nil
	default:
		return "", fmt.Errorf("unknown driver %q", s)
	}
}

type PushEvent struct {
	HeadCommit struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"head_commit"`
	Ref string `mapstructure:"ref"`
}
