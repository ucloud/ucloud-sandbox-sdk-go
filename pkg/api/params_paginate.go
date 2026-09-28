package api

import "net/http"

// PaginateParams is satisfied by a pointer to any params struct that carries a
// nextToken cursor. It lets PageParams move the cursor of whichever listing it
// is handed.
type PaginateParams[T any] interface {
	*T
	SetNextToken(token PaginationNextToken)
}

// PageParams copies params for one request of a listing and points the copy at
// token, the cursor the page before it returned. The caller's struct is never
// written to, so the cursor can move forward page by page. An empty token
// leaves the caller's own cursor in place, so a listing can be resumed from
// one. A nil params yields a zero-valued copy.
func PageParams[T any, P PaginateParams[T]](params *T, token string) *T {
	var query T
	if params != nil {
		query = *params
	}
	if token != "" {
		P(&query).SetNextToken(token)
	}
	return &query
}

func (p *SandboxListParamsV2) SetNextToken(token PaginationNextToken) { p.NextToken = &token }

func (p *SnapshotListParams) SetNextToken(token PaginationNextToken) { p.NextToken = &token }

func (p *SecretListParams) SetNextToken(token PaginationNextToken) { p.NextToken = &token }

func (p *TemplateListParamsV2) SetNextToken(token PaginationNextToken) { p.NextToken = &token }

// Page returns what a paged listing needs from one response: the decoded items
// of the success status, and the raw response and body for the status check
// and the next cursor.
func (r *GetV2SandboxesResponse) Page() (*[]ListedSandbox, *http.Response, []byte) {
	return r.JSON200, r.HTTPResponse, r.Body
}

func (r *GetSnapshotsResponse) Page() (*[]SnapshotInfo, *http.Response, []byte) {
	return r.JSON200, r.HTTPResponse, r.Body
}

func (r *GetSecretsResponse) Page() (*[]Secret, *http.Response, []byte) {
	return r.JSON200, r.HTTPResponse, r.Body
}

func (r *GetV2TemplatesResponse) Page() (*[]Template, *http.Response, []byte) {
	return r.JSON200, r.HTTPResponse, r.Body
}
