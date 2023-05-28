// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package atlascloud

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

type ArchiveFormat string

const (
	// base64 encoded tar format.
	ArchiveFormatB64Tar ArchiveFormat = "B64_TAR"
)

type DirFormat string

const (
	DirFormatAtlas         DirFormat = "ATLAS"
	DirFormatFlyway        DirFormat = "FLYWAY"
	DirFormatGolangMigrate DirFormat = "GOLANG_MIGRATE"
)

// Driver is enum for the field driver
type Driver string

const (
	DriverMysql      Driver = "MYSQL"
	DriverPostgresql Driver = "POSTGRESQL"
	DriverSqlite     Driver = "SQLITE"
	DriverMariadb    Driver = "MARIADB"
)

// Input type of ReportDir
type ReportDirInput struct {
	// Repository full name. e.g., "owner/repo".
	Repo string `json:"repo"`
	// Branch name.
	Branch string `json:"branch"`
	// Commit SHA.
	Commit string `json:"commit"`
	// File path relative to the repository root.
	Path string `json:"path"`
	// The URL back to the action that triggers this upload.
	Url string `json:"url"`
	// Project this directory belongs to.
	Name *string `json:"name"`
	// Atlas driver used to compute directory state.
	Driver Driver `json:"driver"`
	// Directory content.
	Dir string `json:"dir"`
	// Format of the directory.
	DirFormat DirFormat `json:"dirFormat"`
	// Format of the dir archive.
	ArchiveFormat ArchiveFormat `json:"archiveFormat"`
}

// GetRepo returns ReportDirInput.Repo, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetRepo() string { return v.Repo }

// GetBranch returns ReportDirInput.Branch, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetBranch() string { return v.Branch }

// GetCommit returns ReportDirInput.Commit, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetCommit() string { return v.Commit }

// GetPath returns ReportDirInput.Path, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetPath() string { return v.Path }

// GetUrl returns ReportDirInput.Url, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetUrl() string { return v.Url }

// GetName returns ReportDirInput.Name, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetName() *string { return v.Name }

// GetDriver returns ReportDirInput.Driver, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetDriver() Driver { return v.Driver }

// GetDir returns ReportDirInput.Dir, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetDir() string { return v.Dir }

// GetDirFormat returns ReportDirInput.DirFormat, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetDirFormat() DirFormat { return v.DirFormat }

// GetArchiveFormat returns ReportDirInput.ArchiveFormat, and is useful for accessing the field via an interface.
func (v *ReportDirInput) GetArchiveFormat() ArchiveFormat { return v.ArchiveFormat }

// __reportDirInput is used internally by genqlient
type __reportDirInput struct {
	Input ReportDirInput `json:"input"`
}

// GetInput returns __reportDirInput.Input, and is useful for accessing the field via an interface.
func (v *__reportDirInput) GetInput() ReportDirInput { return v.Input }

// reportDirReportDirReportDirPayload includes the requested fields of the GraphQL type ReportDirPayload.
// The GraphQL type's documentation follows.
//
// Return type of ReportDir.
type reportDirReportDirReportDirPayload struct {
	// Indicate if the operation succeeded.
	Success bool `json:"success"`
}

// GetSuccess returns reportDirReportDirReportDirPayload.Success, and is useful for accessing the field via an interface.
func (v *reportDirReportDirReportDirPayload) GetSuccess() bool { return v.Success }

// reportDirResponse is returned by reportDir on success.
type reportDirResponse struct {
	// Report a directory.
	ReportDir reportDirReportDirReportDirPayload `json:"reportDir"`
}

// GetReportDir returns reportDirResponse.ReportDir, and is useful for accessing the field via an interface.
func (v *reportDirResponse) GetReportDir() reportDirReportDirReportDirPayload { return v.ReportDir }

func reportDir(
	ctx context.Context,
	client graphql.Client,
	input ReportDirInput,
) (*reportDirResponse, error) {
	req := &graphql.Request{
		OpName: "reportDir",
		Query: `
mutation reportDir ($input: ReportDirInput!) {
	reportDir(input: $input) {
		success
	}
}
`,
		Variables: &__reportDirInput{
			Input: input,
		},
	}
	var err error

	var data reportDirResponse
	resp := &graphql.Response{Data: &data}

	err = client.MakeRequest(
		ctx,
		req,
		resp,
	)

	return &data, err
}
