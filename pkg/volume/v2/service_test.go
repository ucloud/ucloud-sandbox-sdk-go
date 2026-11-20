package volume_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
)

const testID = "c8fd09b4-a581-4b69-b087-97b97a49ec4f"

const volumeJSON = `{"volumeID":"` + testID + `","name":"test-volume","mountIP":"10.0.0.1","sizeGiB":500,"extraParams":{"storageType":"Basic"},"createdAt":"2026-11-20T01:02:03Z","updatedAt":"2026-11-20T01:02:03Z"}`

func testClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-project", r.Header.Get("X-Project"))
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Options{
		APIKey: "test-key", APIURL: srv.URL + "/gateway", HTTPClient: srv.Client(),
		Headers: map[string]string{"X-Project": "test-project", "X-API-Key": "must-be-overridden"},
	})
	require.NoError(t, err)
	return c
}

func TestCreate(t *testing.T) {
	for _, status := range []int{http.StatusCreated, http.StatusOK} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/gateway/v2/volumes", r.URL.Path)
				assert.Equal(t, "create-test-volume", r.Header.Get("Idempotency-Key"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				var body map[string]any
				if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
					assert.Equal(t, map[string]any{
						"name": "test-volume", "sizeGiB": float64(500),
						"extraParams": map[string]any{"storageType": "Basic"},
					}, body)
				}
				w.WriteHeader(status)
				fmt.Fprint(w, volumeJSON)
			})
			extra := api.ExtraParams{"storageType": "Basic"}
			v, err := c.VolumesV2().Create(context.Background(), api.CreateVolume{
				Name: "test-volume", SizeGiB: 500, ExtraParams: &extra,
			}, &api.CreateVolumeParams{IdempotencyKey: new("create-test-volume")})
			require.NoError(t, err)
			assert.Equal(t, testID, v.VolumeID.String())
			assert.Equal(t, "test-volume", v.Name)
			assert.Equal(t, int64(500), v.SizeGiB)
			assert.Equal(t, "10.0.0.1", v.MountIP)
			assert.Equal(t, extra, v.ExtraParams)
			assert.Equal(t, time.Date(2026, 11, 20, 1, 2, 3, 0, time.UTC), v.CreatedAt)
		})
	}
}

func TestList(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/gateway/v2/volumes", r.URL.Path)
				if explicit {
					assert.Equal(t, "0", r.URL.Query().Get("offset"))
					assert.Equal(t, "10", r.URL.Query().Get("limit"))
				} else {
					assert.Empty(t, r.URL.RawQuery)
				}
				fmt.Fprintf(w, `{"items":[%s],"total":42}`, volumeJSON)
			})
			var params *api.ListVolumesParams
			if explicit {
				params = &api.ListVolumesParams{Offset: new(0), Limit: new(10)}
			}
			page, err := c.VolumesV2().List(context.Background(), params)
			require.NoError(t, err)
			assert.Equal(t, int64(42), page.Total)
			require.Len(t, page.Items, 1)
			assert.Equal(t, testID, page.Items[0].VolumeID.String())
		})
	}
}

func TestGetResizeDelete(t *testing.T) {
	var calls []string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/gateway/v2/volumes/"+testID, r.URL.Path)
		calls = append(calls, r.Method)
		switch r.Method {
		case "GET":
			var detail map[string]any
			if !assert.NoError(t, json.Unmarshal([]byte(volumeJSON), &detail)) {
				return
			}
			detail["totalMountPointNum"] = 3
			assert.NoError(t, json.NewEncoder(w).Encode(detail))
		case "PATCH":
			var body map[string]any
			if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
				assert.Equal(t, map[string]any{"sizeGiB": float64(1000)}, body)
			}
			var v api.Volume
			assert.NoError(t, json.Unmarshal([]byte(volumeJSON), &v))
			v.SizeGiB = 1000
			assert.NoError(t, json.NewEncoder(w).Encode(v))
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})
	ctx := context.Background()
	var id api.VolumeID
	require.NoError(t, id.UnmarshalText([]byte(testID)))
	detail, err := c.VolumesV2().Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, 3, detail.TotalMountPointNum)
	assert.Equal(t, id, detail.VolumeID)
	v, err := c.VolumesV2().Resize(ctx, id, 1000)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), v.SizeGiB)
	require.NoError(t, c.VolumesV2().Delete(ctx, id))
	require.NoError(t, c.VolumesV2().Delete(ctx, id))
	assert.Equal(t, []string{"GET", "PATCH", "DELETE", "DELETE"}, calls)
}

func TestFailures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"invalid argument", 400, `{"code":400,"message":"capacity is too small"}`, errdefs.ErrInvalidArgument},
		{"authentication", 401, `{"code":401,"message":"bad key"}`, errdefs.ErrAuth},
		{"missing", 404, `{"code":404,"message":"missing volume"}`, errdefs.ErrNotFound},
		{"conflict", 409, `{"code":409,"message":"idempotency conflict"}`, errdefs.ErrConflict},
		{"cloud timeout", 504, `{"code":504,"message":"cloud timed out"}`, nil},
		{"empty success", 200, "", nil},
		{"malformed success", 200, "{", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Empty(t, r.Header.Get("Idempotency-Key"))
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			})
			v, err := c.VolumesV2().Create(context.Background(), api.CreateVolume{Name: "test-volume", SizeGiB: 500}, nil)
			require.Error(t, err)
			assert.Nil(t, v)
			if tt.want != nil {
				assert.ErrorIs(t, err, tt.want)
			}
			if tt.status == 504 {
				assert.Contains(t, err.Error(), "HTTP 504: cloud timed out")
			}
			if tt.status >= 400 {
				// Delete must not swallow missing-resource or cloud cleanup errors.
				err = c.VolumesV2().Delete(context.Background(), api.VolumeID{})
				require.Error(t, err)
				if tt.want != nil {
					assert.ErrorIs(t, err, tt.want)
				}
			}
		})
	}
}

func TestCanceledContext(t *testing.T) {
	c := testClient(t, func(http.ResponseWriter, *http.Request) {
		t.Error("canceled request reached server")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.VolumesV2().List(ctx, nil)
	require.True(t, errors.Is(err, context.Canceled), "got %v", err)
}
