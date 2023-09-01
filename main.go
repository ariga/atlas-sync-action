package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"ariga.io/atlas-go-sdk/atlasexec"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/mitchellh/mapstructure"
	"github.com/sethvargo/go-githubactions"

	"ariga.io/atlas-sync-action/internal/atlascloud"
)

const cloudDomainPublic = "https://gh-api.atlasgo.cloud"

func main() {
	act := githubactions.New()
	if ok, err := strconv.ParseBool(act.GetInput("cloud-public")); err == nil && ok {
		RunPublic(act)
	} else {
		RunCmd(act)
	}
	githubactions.Infof("Uploaded migration dir %q to Atlas Cloud", act.GetInput("dir"))
}

// LoadParams loads the atlasexec params from the GitHub Action configuration.
func LoadParams(act *githubactions.Action) (*atlasexec.MigratePushParams, error) {
	c, err := act.Context()
	if err != nil {
		return nil, err
	}
	// Normalize the name.
	reNotSlug := regexp.MustCompile(`[^a-z0-9-._]`)
	name := reNotSlug.ReplaceAllString(
		strings.ToLower(strings.Trim(act.GetInput("dir"), "- \t\n\r")),
		"-",
	)
	org, repo := c.Repo()
	ev := PushEvent{}
	if err := mapstructure.Decode(c.Event, &ev); err != nil {
		return nil, err
	}
	path := act.GetInput("dir")
	dirFmt := act.GetInput("dir-format")
	if err != nil {
		return nil, err
	}
	syncctx := ContextInput{
		Repo:   fmt.Sprintf("%s/%s", org, repo),
		Branch: c.RefName,
		Commit: c.SHA,
		Path:   path,
		URL:    ev.HeadCommit.URL,
	}
	buf, err := json.Marshal(syncctx)
	if err != nil {
		return nil, err
	}
	return &atlasexec.MigratePushParams{
		Name:      name,
		Tag:       act.GetInput("tag"),
		DirURL:    fmt.Sprintf("file://%s", path),
		DevURL:    act.GetInput("dev-url"),
		DirFormat: dirFmt,
		Context:   string(buf),
	}, nil
}

// RunCmd pushed the directory to Atlas Cloud using atlasexec.
func RunCmd(act *githubactions.Action) {
	c, err := atlasexec.NewClient("", "atlas")
	if err != nil {
		act.Fatalf("failed to connect to Atlas Cloud: %v", err)
	}
	if err = c.Login(context.Background(), &atlasexec.LoginParams{Token: act.GetInput("cloud-token")}); err != nil {
		act.Fatalf("failed to login to Atlas Cloud: %v", err)
	}
	params, err := LoadParams(act)
	if err != nil {
		act.Fatalf("failed to load params: %v", err)
	}
	resp, err := c.MigratePush(context.Background(), params)
	if err != nil {
		act.Fatalf("failed to push directory: %v", err)
	}
	act.SetOutput("url", resp)
}

// RunPublic uploads the directory to Atlas Cloud using the public API.
func RunPublic(act *githubactions.Action) {
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
	fi := act.GetInput("dir-format")
	dirFmt, err := dirFormat(fi)
	if err != nil {
		return atlascloud.ReportDirInput{}, err
	}
	return atlascloud.ReportDirInput{
		Name: func() *string {
			if n := act.GetInput("name"); n != "" {
				return &n
			}
			return nil
		}(),
		Repo:          fmt.Sprintf("%s/%s", org, repo),
		Branch:        c.RefName,
		Commit:        c.SHA,
		Path:          act.GetInput("dir"),
		Url:           ev.HeadCommit.URL,
		Driver:        drv,
		DirFormat:     dirFmt,
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

func dirFormat(s string) (atlascloud.DirFormat, error) {
	switch s := strings.ToLower(s); s {
	case "atlas", "":
		return atlascloud.DirFormatAtlas, nil
	case "flyway":
		return atlascloud.DirFormatFlyway, nil
	case "golang-migrate":
		return atlascloud.DirFormatGolangMigrate, nil
	default:
		return "", fmt.Errorf("unknown dir-format %q", s)
	}
}

func client(act *githubactions.Action) *atlascloud.Client {
	token, err := act.GetIDToken(context.Background(), "ariga://atlas-sync-action")
	if err != nil {
		act.Fatalf("failed to get id token: %v", err)
	}
	d := act.GetInput("cloud-url")
	if d == "" {
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

type ContextInput struct {
	Repo   string `json:"repo"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	URL    string `json:"url"`
}
