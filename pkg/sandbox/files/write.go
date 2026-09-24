package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	envdapi "github.com/ucloud/ucloud-sandbox-sdk-go/pkg/envd/api"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/errdefs"
)

// WriteEntry is one file in a multi-file write. Data may be a string, a
// []byte or an io.Reader.
type WriteEntry struct {
	Path string
	Data any
}

// Write writes data to a file, creating any missing parent directories.
//
// data may be a string, a []byte or an io.Reader. Anything else is rejected;
// WriteStream is the statically typed form.
func (f *Filesystem) Write(ctx context.Context, filePath string, data any) (*envdapi.EntryInfo, error) {
	reader, err := asReader(data)
	if err != nil {
		return nil, err
	}
	return f.WriteStream(ctx, filePath, reader)
}

// WriteStream streams reader into a file.
func (f *Filesystem) WriteStream(ctx context.Context, filePath string, reader io.Reader) (*envdapi.EntryInfo, error) {
	written, err := f.WriteFiles(ctx, []WriteEntry{{Path: filePath, Data: reader}})
	if err != nil {
		return nil, err
	}
	if len(written) == 0 {
		return nil, &errdefs.SandboxError{Message: "envd accepted the upload but reported no file"}
	}
	return &written[0], nil
}

// WriteFiles writes several files in one request.
//
// One multipart upload instead of one request per file, which matters when
// seeding a sandbox with a project tree. When a single path is given it is sent
// as the request's path parameter; with several, each part carries its own.
func (f *Filesystem) WriteFiles(ctx context.Context, files []WriteEntry) ([]envdapi.EntryInfo, error) {
	if len(files) == 0 {
		return nil, &errdefs.InvalidArgumentError{SandboxError: errdefs.SandboxError{
			Message: "no files to write",
		}}
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, file := range files {
		reader, err := asReader(file.Data)
		if err != nil {
			return nil, err
		}
		// envd takes each part's target from its filename, so the full path
		// goes there rather than just the base name.
		part, err := writer.CreateFormFile("file", file.Path)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, reader); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	params := &envdapi.PostFilesParams{}
	if user := f.conn.GetUser(); user != "" {
		params.Username = &user
	}
	// With one file the path parameter names it outright; with several, each
	// part's filename does, and a single path here would be ambiguous.
	if len(files) == 1 {
		params.Path = &files[0].Path
	}

	resp, err := f.conn.Files.PostFilesWithBody(ctx, params, writer.FormDataContentType(), &body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkEnvdResponse(resp); err != nil {
		return nil, err
	}

	var entries []envdapi.EntryInfo
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// Compose concatenates files into one, without copying their bytes through
// this process. The sources are removed once the destination is written.
func (f *Filesystem) Compose(ctx context.Context, destination string, sources []string) (*envdapi.EntryInfo, error) {
	body := envdapi.PostFilesComposeJSONRequestBody{
		Destination: destination,
		SourcePaths: sources,
	}
	if user := f.conn.GetUser(); user != "" {
		body.Username = &user
	}

	resp, err := f.conn.Files.PostFilesComposeWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := checkEnvdResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &errdefs.SandboxError{Message: "envd composed the files but reported no entry"}
	}

	return resp.JSON200, nil
}

// asReader accepts the three shapes Write takes.
func asReader(data any) (io.Reader, error) {
	switch value := data.(type) {
	case string:
		return strings.NewReader(value), nil
	case []byte:
		return bytes.NewReader(value), nil
	case io.Reader:
		return value, nil
	default:
		return nil, &errdefs.InvalidArgumentError{SandboxError: errdefs.SandboxError{
			Message: fmt.Sprintf("cannot write %T to a file: pass a string, []byte or io.Reader", data),
		}}
	}
}
