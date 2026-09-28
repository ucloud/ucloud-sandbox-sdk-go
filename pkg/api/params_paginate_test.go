package api

import "testing"

func TestPageParams(t *testing.T) {
	if got := PageParams[SecretListParams](nil, ""); got == nil || got.NextToken != nil {
		t.Fatalf("nil params: got %+v", got)
	}

	resume := "resume"
	params := &SecretListParams{NextToken: &resume, Limit: new(PaginationLimit(10))}

	got := PageParams(params, "")
	if got == params || *got.NextToken != "resume" || *got.Limit != 10 {
		t.Fatalf("empty token: got %+v", got)
	}

	got = PageParams(params, "next")
	if *got.NextToken != "next" || *got.Limit != 10 {
		t.Fatalf("with token: got %+v", got)
	}
	if *params.NextToken != "resume" {
		t.Fatalf("caller's params were written to: %q", *params.NextToken)
	}
}
