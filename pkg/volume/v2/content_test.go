package volume_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/client"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
	volume "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/volume/v2/contentapi"
)

const entryJSON = `{"name":"a.txt","type":"file","path":"/dir/a.txt","size":5,"mode":420,"uid":1000,"gid":1000,"atime":"2026-11-20T01:02:03Z","mtime":"2026-11-20T01:02:03Z","ctime":"2026-11-20T01:02:03Z"}`

const contentPrefix = "/content/v2/volumecontent/" + testID

// testContent serves the content API under /content and fails any request
// that reaches the control-plane URL instead.
func testContent(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.URL.Path, "/content/"), "request went to %s", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-project", r.Header.Get("X-Project"))
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Options{
		APIKey: "test-key", APIURL: srv.URL + "/gateway", VolumeContentURL: srv.URL + "/content/",
		HTTPClient: srv.Client(),
		Headers:    map[string]string{"X-Project": "test-project"},
	})
	require.NoError(t, err)
	return c
}

func volumeID(t *testing.T) api.VolumeID {
	t.Helper()
	var id api.VolumeID
	require.NoError(t, id.UnmarshalText([]byte(testID)))
	return id
}

func TestContentDefaultHost(t *testing.T) {
	c, err := client.New(client.Options{APIKey: "k", Region: "cn-wlcb"})
	require.NoError(t, err)
	assert.Equal(t, "https://volumecontent.cn-wlcb.sandbox.ucloudai.com", c.Transport().VolumeContentURL())
}

func TestContentReadFile(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, contentPrefix+"/file", r.URL.Path)
		assert.Equal(t, "/dir/a b.txt", r.URL.Query().Get("path"))
		w.Header().Set("Content-Type", "application/octet-stream")
		fmt.Fprint(w, "hello")
	})
	rc, err := c.VolumeContentV2(volumeID(t)).ReadFile(context.Background(), "/dir/a b.txt")
	require.NoError(t, err)
	defer rc.Close()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestContentReadFileError(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"code":404,"message":"no such file"}`)
	})
	_, err := c.VolumeContentV2(volumeID(t)).ReadFile(context.Background(), "/missing")
	var nf *errdefs.NotFoundError
	require.True(t, errors.As(err, &nf), "got %T: %v", err, err)
	assert.Contains(t, err.Error(), "no such file")
}

func TestContentWriteFile(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, contentPrefix+"/file", r.URL.Path)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
		q := r.URL.Query()
		assert.Equal(t, "/dir/a.txt", q.Get("path"))
		assert.Equal(t, "true", q.Get("force"))
		assert.Equal(t, "1000", q.Get("uid"))
		assert.False(t, q.Has("gid"))
		assert.Equal(t, "420", q.Get("mode"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "hello", string(body))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, entryJSON)
	})
	e, err := c.VolumeContentV2(volumeID(t)).WriteFile(context.Background(), "/dir/a.txt",
		bytes.NewReader([]byte("hello")), &volume.CreateOptions{Force: true, UID: new(int64(1000)), Mode: new(fs.FileMode(0o644))})
	require.NoError(t, err)
	assert.Equal(t, contentapi.EntryType("file"), e.Type)
	assert.Equal(t, int64(5), e.Size)
	assert.Equal(t, 0o644, e.Mode)
}

func TestContentWriteFileNilOptions(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "path=%2Fa.txt", r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, entryJSON)
	})
	_, err := c.VolumeContentV2(volumeID(t)).WriteFile(context.Background(), "/a.txt", strings.NewReader("x"), nil)
	require.NoError(t, err)
}

func TestContentMakeDir(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, contentPrefix+"/dir", r.URL.Path)
		assert.Equal(t, "/dir", r.URL.Query().Get("path"))
		assert.Equal(t, "493", r.URL.Query().Get("mode"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, strings.Replace(entryJSON, `"type":"file"`, `"type":"directory"`, 1))
	})
	e, err := c.VolumeContentV2(volumeID(t)).MakeDir(context.Background(), "/dir",
		&volume.CreateOptions{Mode: new(fs.FileMode(0o755))})
	require.NoError(t, err)
	assert.Equal(t, contentapi.EntryType("directory"), e.Type)
}

func TestContentListDir(t *testing.T) {
	for _, depth := range []int{0, 3} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, contentPrefix+"/dir", r.URL.Path)
				assert.Equal(t, "/", r.URL.Query().Get("path"))
				if depth == 0 {
					assert.False(t, r.URL.Query().Has("depth"))
				} else {
					assert.Equal(t, "3", r.URL.Query().Get("depth"))
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, "["+entryJSON+"]")
			})
			entries, err := c.VolumeContentV2(volumeID(t)).ListDir(context.Background(), "/", depth)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, "a.txt", entries[0].Name)
		})
	}
}

func TestContentStatAndExists(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, contentPrefix+"/path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("path") {
		case "/dir/a.txt":
			fmt.Fprint(w, entryJSON)
		case "/boom":
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"code":500,"message":"boom"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"code":404,"message":"not found"}`)
		}
	})
	vc := c.VolumeContentV2(volumeID(t))
	ctx := context.Background()

	e, err := vc.Stat(ctx, "/dir/a.txt")
	require.NoError(t, err)
	assert.Equal(t, "/dir/a.txt", e.Path)

	ok, err := vc.Exists(ctx, "/dir/a.txt")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = vc.Exists(ctx, "/missing")
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = vc.Exists(ctx, "/boom")
	require.Error(t, err)
}

func TestContentUpdateMetadata(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, contentPrefix+"/path", r.URL.Path)
		assert.Equal(t, "/dir/a.txt", r.URL.Query().Get("path"))
		var body map[string]any
		if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			assert.Equal(t, map[string]any{"mode": float64(384)}, body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, entryJSON)
	})
	_, err := c.VolumeContentV2(volumeID(t)).UpdateMetadata(context.Background(), "/dir/a.txt",
		contentapi.Metadata{Mode: new(0o600)})
	require.NoError(t, err)
}

func TestContentRename(t *testing.T) {
	for _, tc := range []struct {
		name  string
		force *bool
	}{
		{name: "default force"},
		{name: "force false", force: new(false)},
		{name: "force true", force: new(true)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const oldPath = "/old/a +&?.txt"
			const newPath = "/dir/新文件 +&?.txt"
			c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, contentPrefix+"/rename", r.URL.Path)
				q := r.URL.Query()
				assert.Equal(t, oldPath, q.Get("path"))
				assert.Equal(t, newPath, q.Get("newPath"))
				if tc.force != nil {
					assert.Equal(t, fmt.Sprint(*tc.force), q.Get("force"))
				} else {
					assert.False(t, q.Has("force"))
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, strings.NewReplacer(
					`"name":"a.txt"`, `"name":"新文件 +&?.txt"`,
					`"path":"/dir/a.txt"`, `"path":"`+newPath+`"`,
				).Replace(entryJSON))
			})
			e, err := c.VolumeContentV2(volumeID(t)).Rename(context.Background(), &contentapi.RenameVolumePathParams{
				Path: oldPath, NewPath: newPath, Force: tc.force,
			})
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.Equal(t, newPath, e.Path)
			assert.Equal(t, "新文件 +&?.txt", e.Name)
			assert.Equal(t, contentapi.EntryType("file"), e.Type)
			assert.Equal(t, int64(5), e.Size)
		})
	}
}

func TestContentRenameError(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusConflict} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			c := testContent(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				fmt.Fprintf(w, `{"code":%d,"message":"rename failed"}`, status)
			})
			e, err := c.VolumeContentV2(volumeID(t)).Rename(context.Background(), &contentapi.RenameVolumePathParams{
				Path: "/old", NewPath: "/new",
			})
			require.Error(t, err)
			assert.Nil(t, e)
			assert.Contains(t, err.Error(), "rename failed")
			if status == http.StatusNotFound {
				assert.ErrorIs(t, err, errdefs.ErrNotFound)
			} else {
				assert.ErrorIs(t, err, errdefs.ErrConflict)
			}
		})
	}
}

func TestContentRemove(t *testing.T) {
	c := testContent(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, contentPrefix+"/path", r.URL.Path)
		assert.Equal(t, "/dir", r.URL.Query().Get("path"))
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, c.VolumeContentV2(volumeID(t)).Remove(context.Background(), "/dir"))
}
