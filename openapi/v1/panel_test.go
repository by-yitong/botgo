package v1

// 本地 fork 补充的指令面板详情/修改/关联对象测试（官方 2026-11 文档新增）：
// 经真实 openAPI 实例打到 httptest 假 server（覆盖 constant.APIDomain 指向本机）。

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/by-yitong/botgo/dto"
)

func TestGetPanel(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		fmt.Fprint(w, `{"panel_id":"P1","scope":"group","target_type":"specific",
			"panel":{"items":[{"name":"签到","desc":"每日签到","type":"command"}],"remark":"r"},
			"version":3,"group_openids":["G1","G2"],"created_at":"2026-11-01T00:00:00+08:00"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	d, err := api.GetPanel(context.Background(), "P1")
	if err != nil {
		t.Fatalf("GetPanel: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/panels/P1") {
		t.Errorf("path=%q want /v2/panels/P1", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method=%q want GET", gotMethod)
	}
	if d.PanelID != "P1" || d.Scope != "group" || d.TargetType != "specific" || d.Version != 3 {
		t.Errorf("PanelDetail 基础字段错: %+v", d)
	}
	if len(d.Panel.Items) != 1 || d.Panel.Items[0].Name != "签到" || d.Panel.Items[0].Type != "command" {
		t.Errorf("panel.items 应反序列化: %+v", d.Panel.Items)
	}
	if d.Panel.Remark != "r" {
		t.Errorf("panel.remark=%q", d.Panel.Remark)
	}
	if len(d.GroupOpenIDs) != 2 || d.GroupOpenIDs[0] != "G1" || d.GroupOpenIDs[1] != "G2" {
		t.Errorf("group_openids 应反序列化: %v", d.GroupOpenIDs)
	}
}

func TestUpdatePanel(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"version":4}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	rsp, err := api.UpdatePanel(context.Background(), "P1", dto.Panel{
		Items:  []dto.PanelItem{{Name: "帮助", Type: "command"}},
		Remark: "新备注",
	})
	if err != nil {
		t.Fatalf("UpdatePanel: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/panels/P1") {
		t.Errorf("path=%q want /v2/panels/P1", gotPath)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method=%q want PUT", gotMethod)
	}
	var req struct {
		Panel struct {
			Items  []dto.PanelItem `json:"items"`
			Remark string          `json:"remark"`
		} `json:"panel"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if len(req.Panel.Items) != 1 || req.Panel.Items[0].Name != "帮助" || req.Panel.Remark != "新备注" {
		t.Errorf("请求体 panel 字段错: %+v", req.Panel)
	}
	if rsp.Version != 4 {
		t.Errorf("响应 version=%d want 4", rsp.Version)
	}
}

func TestUpdatePanelTarget(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if err := api.UpdatePanelTarget(context.Background(), "P1", dto.PanelTargetReq{
		Op:           "add",
		GroupOpenIDs: []string{"G1", "G2"},
	}); err != nil {
		t.Fatalf("UpdatePanelTarget: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/panels/P1/target") {
		t.Errorf("path=%q want /v2/panels/P1/target", gotPath)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method=%q want PUT", gotMethod)
	}
	var req dto.PanelTargetReq
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if req.Op != "add" || len(req.GroupOpenIDs) != 2 || req.GroupOpenIDs[1] != "G2" {
		t.Errorf("请求体 op/group_openids 错: %+v", req)
	}
}

func TestGetPanelHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"code":40030006,"message":"面板不存在"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if _, err := api.GetPanel(context.Background(), "NOPE"); err == nil {
		t.Fatal("HTTP 404（40030006 面板不存在）应报错")
	}
}
