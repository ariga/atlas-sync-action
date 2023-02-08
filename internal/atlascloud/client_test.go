package atlascloud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	var (
		success bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer atlas", r.Header.Get("Authorization"))
		require.Equal(t, UserAgent, r.Header.Get("User-Agent"))
		fmt.Fprintf(w, `{"data":{"reportDir":{"success":%v}}}`, success)
	}))
	client := New(srv.URL, "atlas")
	defer srv.Close()
	t.Run("success", func(t *testing.T) {
		success = true
		err := client.ReportDir(context.Background(), ReportDirInput{})
		require.NoError(t, err)
	})
	t.Run("fail", func(t *testing.T) {
		success = false
		err := client.ReportDir(context.Background(), ReportDirInput{})
		require.EqualError(t, err, "upload failed")
	})
}
