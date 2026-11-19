package v1

// 本地 fork 补充的生成分享链接测试（官方 2026-11 文档新增）：
// 经真实 openAPI 实例打到 httptest 假 server（覆盖 constant.APIDomain 指向本机）。

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateURLLink(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"data":{"url":"https://q.qq.com/s/abc123"}}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	url, err := api.GenerateURLLink(context.Background(), "cb-data-1")
	if err != nil {
		t.Fatalf("GenerateURLLink: %v", err)
	}
	if url != "https://q.qq.com/s/abc123" {
		t.Errorf("url=%q want https://q.qq.com/s/abc123", url)
	}
	if !strings.HasSuffix(gotPath, "/v2/generate_url_link") {
		t.Errorf("path=%q want /v2/generate_url_link", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method=%q want POST", gotMethod)
	}
	if !strings.Contains(string(bodyBytes), `"callback_data":"cb-data-1"`) {
		t.Errorf("请求体应带 callback_data: %q", bodyBytes)
	}
}

func TestGenerateURLLinkEmptyCallback(t *testing.T) {
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"data":{"url":"https://q.qq.com/s/empty"}}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	url, err := api.GenerateURLLink(context.Background(), "")
	if err != nil {
		t.Fatalf("GenerateURLLink: %v", err)
	}
	if url != "https://q.qq.com/s/empty" {
		t.Errorf("url=%q", url)
	}
	if trimmed := strings.TrimSpace(string(bodyBytes)); trimmed != "{}" {
		t.Errorf("callback_data 为空时请求体应为空对象: %q", trimmed)
	}
}

func TestGenerateURLLinkHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"code":11004,"message":"生成分享ARK失败"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if _, err := api.GenerateURLLink(context.Background(), "x"); err == nil {
		t.Fatal("HTTP 400（11004 生成分享 ARK 失败）应报错")
	}
}
