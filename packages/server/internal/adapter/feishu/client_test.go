package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testServer(t *testing.T, check func(*http.Request), response map[string]any) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check(r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	return server, NewClient(server.URL, "test-token", server.Client())
}

func TestBitableMethods(t *testing.T) {
	calls := 0
	server, client := testServer(t, func(r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header")
		}
	}, map[string]any{"code": 0, "data": map[string]any{"ok": true}})
	defer server.Close()
	ctx := context.Background()
	if _, err := client.CreateApp(ctx, "demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateField(ctx, "app", "tbl", "name", 1, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateRecord(ctx, "app", "tbl", map[string]any{"name": "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRecord(ctx, "app", "tbl", "rec"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListRecords(ctx, "app", "tbl", 10, "next"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateRecord(ctx, "app", "tbl", "rec", map[string]any{"name": "y"}); err != nil {
		t.Fatal(err)
	}
	if calls != 6 {
		t.Fatalf("calls=%d, want 6", calls)
	}
}

func TestDocMethods(t *testing.T) {
	server, client := testServer(t, func(r *http.Request) {}, map[string]any{"code": 0, "data": map[string]any{"ok": true}})
	defer server.Close()
	ctx := context.Background()
	if _, err := client.Read(ctx, "doc"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(ctx, "doc", []map[string]any{{"block_type": 2}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Append(ctx, "doc", []map[string]any{{"block_type": 2}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateTable(ctx, "doc", 2, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WriteTableCells(ctx, "doc", "tbl", [][]string{{"a"}}); err != nil {
		t.Fatal(err)
	}
}

func TestDriveMethods(t *testing.T) {
	seenQuery := false
	server, client := testServer(t, func(r *http.Request) {
		if r.URL.Path == "/open-apis/drive/v1/files" {
			seenQuery = r.URL.Query().Get("folder_token") == "folder" && r.URL.Query().Get("page_size") == "10"
		}
	}, map[string]any{"code": 0, "data": map[string]any{"ok": true}})
	defer server.Close()
	ctx := context.Background()
	if _, err := client.List(ctx, "folder", 10, "page"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Info(ctx, "file"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateFolder(ctx, "name", "parent"); err != nil {
		t.Fatal(err)
	}
	if !seenQuery {
		t.Fatal("list query was not encoded")
	}
}

func TestWikiMethods(t *testing.T) {
	server, client := testServer(t, func(r *http.Request) {}, map[string]any{"code": 0, "data": map[string]any{"ok": true}})
	defer server.Close()
	ctx := context.Background()
	if _, err := client.Spaces(ctx, 10, "page"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Nodes(ctx, "space", "parent", 10, "page"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get(ctx, "node"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Create(ctx, "space", "title", "parent"); err != nil {
		t.Fatal(err)
	}
}

func TestClientErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":91403,"msg":"forbidden"}`))
	}))
	defer server.Close()
	_, err := NewClient(server.URL, "", server.Client()).Get(context.Background(), "node")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.HTTPStatus != http.StatusForbidden || apiErr.Code != 91403 {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientRejectsMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()
	_, err := NewClient(server.URL, "", server.Client()).Get(context.Background(), "node")
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestPageQuery(t *testing.T) {
	q := pageQuery(20, "next")
	if q.Get("page_size") != "20" || q.Get("page_token") != "next" {
		t.Fatalf("unexpected query: %v", q)
	}
	if got := pageQuery(0, ""); len(got) != 0 {
		t.Fatalf("expected empty query, got %v", got)
	}
	if _, err := url.ParseQuery(q.Encode()); err != nil {
		t.Fatal(err)
	}
}
