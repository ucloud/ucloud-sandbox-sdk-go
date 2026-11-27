package content_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	volumeapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/content"
	contentapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/content/api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestWriteFileUsesConfiguredIPVolumeContentHostAndGeneratedParams(t *testing.T) {
	const volumeID = "550e8400-e29b-41d4-a716-446655440000"
	var got *http.Request
	var gotBody string
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		gotBody = string(bodyBytes)
		body := `{"name":"note.txt","type":"file","path":"note.txt","size":5,"mode":438,"uid":9090,"gid":9090,"modifiedTime":"2026-11-27T00:00:00Z"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})}
	transportClient, err := transport.New(transport.Config{
		APIKey: "test-key", Region: "us-ca", VolumeContentURL: "http://10.0.0.5:8081/", HTTPClient: httpClient,
	})
	require.NoError(t, err)
	id := volumeapi.VolumeID(uuid.MustParse(volumeID))
	volumeContent, err := content.NewWithTransport(id, transportClient)
	require.NoError(t, err)

	entry, err := volumeContent.WriteFile(context.Background(), &contentapi.WriteVolumeFileParams{
		Path: "note.txt",
	}, strings.NewReader("hello"))
	require.NoError(t, err)
	require.Equal(t, "note.txt", entry.Path)
	require.Equal(t, "10.0.0.5:8081", got.URL.Host)
	require.Equal(t, "/"+volumeID+"/files", got.URL.Path)
	require.Equal(t, "note.txt", got.URL.Query().Get("path"))
	require.Equal(t, "test-key", got.Header.Get("X-API-Key"))
	require.Equal(t, "application/octet-stream", got.Header.Get("Content-Type"))
	require.Equal(t, "hello", gotBody)
}
