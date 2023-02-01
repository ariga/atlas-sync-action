package atlascloud

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
)

const (
	UserAgent = "atlas-sync-action"
)

// Client is a client for the Atlas Cloud API.
type Client struct {
	client graphql.Client
}

// New creates a new Client for the Atlas Cloud API.
func New(endpoint, token string) *Client {
	c := graphql.NewClient(endpoint, &http.Client{
		Transport: &roundTripper{
			token: token,
		},
		Timeout: time.Second * 30,
	})
	return &Client{client: c}
}

// UploadDir uploads a directory to the Atlas Cloud API.
func (c *Client) UploadDir(ctx context.Context, input UploadDirInput) error {
	_ = `# @genqlient
	mutation uploadDir($input: UploadDirInput!) {
		uploadDir(input: $input) {
			success
		}
	}`
	p, err := uploadDir(ctx, c.client, input)
	if err != nil {
		return err
	}
	if !p.GetUploadDir().Success {
		return errors.New("upload failed")
	}
	return nil
}

// roundTripper is a http.RoundTripper that adds the authorization header.
type roundTripper struct {
	token string
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultTransport.RoundTrip(req)
}
