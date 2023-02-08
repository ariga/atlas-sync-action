package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"ariga.io/atlas-sync-action/internal/atlascloud"
	"ariga.io/atlas/sql/migrate"
	"github.com/mitchellh/mapstructure"
	"github.com/sethvargo/go-githubactions"
)

const (
	cloudDomain = "https://ingress.atlasgo.cloud"
)

func main() {
	act := githubactions.New()
	token := act.GetInput("cloud-token")
	if token == "" {
		act.Fatalf("cloud-token is required")
	}
	input, err := Input(act)
	if err != nil {
		act.Fatalf("failed to parse input: %v", err)
	}
	arc, err := Archive(input.Path)
	if err != nil {
		act.Fatalf("failed to archive migration dir: %v", err)
	}
	input.Dir = arc
	c := client(act)
	if err := c.UploadDir(context.Background(), input); err != nil {
		act.Fatalf("failed to upload dir: %v", err)
	}
	githubactions.Infof("Uploaded migration dir %q to Atlas Cloud", input.Path)
}

type file struct {
	name string
	data []byte
}

// Archive returns a b64 encoded tarball of the given migration directory.
func Archive(path string) (string, error) {
	d, err := migrate.NewLocalDir(path)
	if err != nil {
		return "", err
	}
	files, err := d.Files()
	if err != nil {
		return "", err
	}
	var arc []file
	for _, f := range files {
		arc = append(arc, file{
			name: f.Name(),
			data: f.Bytes(),
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
	arc = append(arc, file{
		name: migrate.HashFileName,
		data: wantS,
	})
	return b64tar(arc)
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

func client(act *githubactions.Action) *atlascloud.Client {
	token := act.GetInput("cloud-token")
	if token == "" {
		act.Fatalf("cloud-token is required")
	}
	d := cloudDomain
	if u := act.GetInput("cloud-url"); u != "" {
		d = u
	}
	u, err := url.Parse(d)
	if err != nil {
		act.Fatalf("failed to parse cloud-url: %v", err)
	}
	u.Path = "/api/query"
	return atlascloud.New(u.String(), token)
}

type PushEvent struct {
	HeadCommit struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"head_commit"`
	Ref string `mapstructure:"ref"`
}

func b64tar(files []file) (string, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, f := range files {
		hdr := &tar.Header{
			Name: f.name,
			Mode: 0600,
			Size: int64(len(f.data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", err
		}
		if _, err := tw.Write(f.data); err != nil {
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
