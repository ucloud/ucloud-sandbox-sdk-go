package volume_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/transport"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume"
	volumeapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestCreateAndResizeUseExtendedVolumeAPI(t *testing.T) {
	const volumeID = "550e8400-e29b-41d4-a716-446655440000"
	var requests []*http.Request
	var requestBodies []string

	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req)
		requestBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		requestBodies = append(requestBodies, string(requestBody))
		var body string
		status := http.StatusOK
		if req.URL.Path == "/volumes" {
			status = http.StatusCreated
			body = `{"volumeID":"` + volumeID + `","name":"project-data","totalSizeGiB":10,"usedSizeGiB":0,"availableSizeGiB":10}`
		} else {
			body = `{"volumeID":"` + volumeID + `","name":"project-data","totalSizeGiB":20,"usedSizeGiB":0,"availableSizeGiB":20}`
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})}
	transportClient, err := transport.New(transport.Config{APIKey: "test-key", APIURL: "https://api.example.test", HTTPClient: httpClient})
	require.NoError(t, err)
	svc, err := volume.NewService(transportClient)
	require.NoError(t, err)

	size := int64(10)
	created, err := svc.Create(context.Background(), volumeapi.NewVolume{
		Name: "project-data", TotalSizeGiB: &size,
	})
	require.NoError(t, err)
	require.Equal(t, volumeID, created.VolumeID.String())
	require.Equal(t, int64(10), *created.TotalSizeGiB)

	resized, err := svc.Resize(context.Background(), created.VolumeID, volumeapi.ResizeVolumeRequest{
		TotalSizeGiB: 20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(20), *resized.TotalSizeGiB)

	require.Len(t, requests, 2)
	require.Equal(t, "test-key", requests[0].Header.Get("X-API-Key"))
	require.Equal(t, "/volumes", requests[0].URL.Path)
	require.JSONEq(t, `{"name":"project-data","totalSizeGiB":10}`, requestBodies[0])
	require.Equal(t, "/volumes/"+volumeID+"/resize", requests[1].URL.Path)
	require.JSONEq(t, `{"totalSizeGiB":20}`, requestBodies[1])
}
