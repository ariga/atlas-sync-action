package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/mitchellh/mapstructure"
	"github.com/sethvargo/go-githubactions"

	"ariga.io/atlas-sync-action/internal/atlascloud"
)

const (
	cloudDomain       = "https://api.atlasgo.cloud"
	cloudDomainPublic = "https://gh-api.atlasgo.cloud"
)

func main() {
	act := githubactions.New()
	c := client(act)
	input, err := Input(act)
	if err != nil {
		act.Fatalf("failed to parse input: %v", err)
	}
	arc, err := Archive(input.Path, input.DirFormat)
	if err != nil {
		act.Fatalf("failed to archive migration dir: %v", err)
	}
	input.Dir = arc
	if err := c.ReportDir(context.Background(), input); err != nil {
		act.Fatalf("failed to upload dir: %v", err)
	}
	githubactions.Infof("Uploaded migration dir %q to Atlas Cloud", input.Path)
}

// Archive returns a b64 encoded tarball of the given migration directory.
func Archive(path string, format atlascloud.DirFormat) (string, error) {
	var (
		dir migrate.Dir
		err error
	)
	switch format {
	case atlascloud.DirFormatAtlas:
		dir, err = migrate.NewLocalDir(path)
	case atlascloud.DirFormatFlyway:
		dir, err = sqltool.NewFlywayDir(path)
	case atlascloud.DirFormatGolangMigrate:
		dir, err = sqltool.NewGolangMigrateDir(path)
	default:
		return "", fmt.Errorf("unknown dir format %q", format)
	}
	if err != nil {
		return "", err
	}
	arc, err := migrate.ArchiveDir(dir)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(arc), nil
}

func Input(act *githubactions.Action) (atlascloud.ReportDirInput, error) {
	c, err := act.Context()
	if err != nil {
		return atlascloud.ReportDirInput{}, err
	}
	org, repo := c.Repo()
	ev := PushEvent{}
	if err := mapstructure.Decode(c.Event, &ev); err != nil {
		return atlascloud.ReportDirInput{}, err
	}
	di := act.GetInput("driver")
	drv, err := driver(di)
	if err != nil {
		return atlascloud.ReportDirInput{}, err
	}
	dirFormat := atlascloud.DirFormat(strings.ToUpper(act.GetInput("dir-format")))
	if dirFormat == "" {
		dirFormat = atlascloud.DirFormatAtlas
	}
	return atlascloud.ReportDirInput{
		Repo:          fmt.Sprintf("%s/%s", org, repo),
		Branch:        c.RefName,
		Commit:        c.SHA,
		Path:          act.GetInput("dir"),
		Url:           ev.HeadCommit.URL,
		Driver:        drv,
		DirFormat:     dirFormat,
		ArchiveFormat: atlascloud.ArchiveFormatB64Tar,
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
	case "maria", "mariadb":
		return atlascloud.DriverMariadb, nil
	default:
		return "", fmt.Errorf("unknown driver %q", s)
	}
}

func client(act *githubactions.Action) *atlascloud.Client {
	isPublic := strings.ToLower(act.GetInput("cloud-public")) == "true"
	token := act.GetInput("cloud-token")
	if token == "" && isPublic {
		var err error
		token, err = act.GetIDToken(context.Background(), "ariga://atlas-sync-action")
		if err != nil {
			act.Fatalf("failed to get id token: %v", err)
		}
	}
	if token == "" {
		act.Fatalf("cloud-token is required")
	}
	d := cloudDomain
	switch u := act.GetInput("cloud-url"); {
	case u != "":
		d = u
	case isPublic:
		d = cloudDomainPublic
	}
	u, err := url.Parse(d)
	if err != nil {
		act.Fatalf("failed to parse cloud-url: %v", err)
	}
	u.Path = "/query"
	return atlascloud.New(u.String(), token)
}

type PushEvent struct {
	HeadCommit struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"head_commit"`
	Ref string `mapstructure:"ref"`
}
