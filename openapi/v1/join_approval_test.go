package v1

// 本地 fork 补充的入群自动审批策略管理测试（官方 2026-11 文档新增）：
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

func TestCreateJoinApprovalStrategy(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"strategy_id":"S1","is_enable":"on","expire_at":"2027-11-19T00:00:00+08:00"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	rsp, err := api.CreateJoinApprovalStrategy(context.Background(), dto.CreateJoinApprovalStrategyReq{
		GroupOpenIDs: []string{"G1", "G2"},
		IsEnable:     "on",
		ExpireAt:     "2027-11-19T00:00:00+08:00",
		Remark:       "老用户秒过",
	})
	if err != nil {
		t.Fatalf("CreateJoinApprovalStrategy: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/groups/join_approval_strategy") {
		t.Errorf("path=%q want /v2/groups/join_approval_strategy", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method=%q want POST", gotMethod)
	}
	var req struct {
		GroupOpenIDs []string `json:"group_openids"`
		GroupIDs     []uint64 `json:"group_ids"`
		IsEnable     string   `json:"is_enable"`
		ExpireAt     string   `json:"expire_at"`
		Remark       string   `json:"remark"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if len(req.GroupOpenIDs) != 2 || req.GroupOpenIDs[0] != "G1" || req.IsEnable != "on" ||
		req.ExpireAt != "2027-11-19T00:00:00+08:00" || req.Remark != "老用户秒过" || req.GroupIDs != nil {
		t.Errorf("创建请求体字段错: %+v", req)
	}
	if rsp.StrategyID != "S1" || rsp.IsEnable != "on" || rsp.ExpireAt != "2027-11-19T00:00:00+08:00" {
		t.Errorf("响应字段错: %+v", rsp)
	}
}

func TestUpdateJoinApprovalStrategy(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"strategy_id":"S1","is_enable":"off"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	rsp, err := api.UpdateJoinApprovalStrategy(context.Background(), "S1", dto.UpdateJoinApprovalStrategyReq{
		IsEnable: "off",
		GroupAction: &dto.JoinGroupAction{
			Op:       "del",
			GroupIDs: []uint64{123456},
		},
		Remark: "先停用",
	})
	if err != nil {
		t.Fatalf("UpdateJoinApprovalStrategy: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/groups/join_approval_strategy/S1") {
		t.Errorf("path=%q want /v2/groups/join_approval_strategy/S1", gotPath)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method=%q want PATCH", gotMethod)
	}
	var req struct {
		IsEnable    string `json:"is_enable"`
		GroupAction *struct {
			Op       string   `json:"op"`
			GroupIDs []uint64 `json:"group_ids"`
		} `json:"group_action"`
		Remark string `json:"remark"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if req.IsEnable != "off" || req.Remark != "先停用" || req.GroupAction == nil ||
		req.GroupAction.Op != "del" || len(req.GroupAction.GroupIDs) != 1 || req.GroupAction.GroupIDs[0] != 123456 {
		t.Errorf("修改请求体字段错: %+v", req)
	}
	if rsp.StrategyID != "S1" || rsp.IsEnable != "off" {
		t.Errorf("响应字段错: %+v", rsp)
	}
}

func TestExecuteJoinApprovalStrategy(t *testing.T) {
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
	if err := api.ExecuteJoinApprovalStrategy(context.Background(), "S1"); err != nil {
		t.Fatalf("ExecuteJoinApprovalStrategy: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/groups/join_approval_strategy/S1/execute") {
		t.Errorf("path=%q want /v2/groups/join_approval_strategy/S1/execute", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method=%q want POST", gotMethod)
	}
	if trimmed := strings.TrimSpace(string(bodyBytes)); trimmed != "{}" {
		t.Errorf("执行请求体应为空对象: %q", trimmed)
	}
}

func TestSetJoinApprovalWhitelist(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"strategy_id":"S1","whitelist_user_count":2,"updated_at":"2026-11-19T10:00:00+08:00"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	rsp, err := api.SetJoinApprovalWhitelist(context.Background(), "S1", dto.SetJoinApprovalWhitelistReq{
		Op:             "add",
		WhitelistUsers: []string{"10001", "10002"},
	})
	if err != nil {
		t.Fatalf("SetJoinApprovalWhitelist: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/groups/join_approval_strategy/S1/whitelist_users") {
		t.Errorf("path=%q want /v2/groups/join_approval_strategy/S1/whitelist_users", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method=%q want POST", gotMethod)
	}
	var req struct {
		Op             string   `json:"op"`
		WhitelistUsers []string `json:"whitelist_users"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if req.Op != "add" || len(req.WhitelistUsers) != 2 || req.WhitelistUsers[1] != "10002" {
		t.Errorf("白名单请求体字段错: %+v", req)
	}
	if rsp.StrategyID != "S1" || rsp.WhitelistUserCount != 2 || rsp.UpdatedAt != "2026-11-19T10:00:00+08:00" {
		t.Errorf("响应字段错: %+v", rsp)
	}
}

func TestDeleteJoinApprovalStrategy(t *testing.T) {
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
	if err := api.DeleteJoinApprovalStrategy(context.Background(), "S1"); err != nil {
		t.Fatalf("DeleteJoinApprovalStrategy: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/groups/join_approval_strategy/S1") {
		t.Errorf("path=%q want /v2/groups/join_approval_strategy/S1", gotPath)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method=%q want DELETE", gotMethod)
	}
	if trimmed := strings.TrimSpace(string(bodyBytes)); trimmed != "" {
		t.Errorf("删除不应发请求体: %q", trimmed)
	}
}

func TestDeleteJoinApprovalStrategyEmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if err := api.DeleteJoinApprovalStrategy(context.Background(), "S1"); err != nil {
		t.Fatalf("空响应体不应解析报错: %v", err)
	}
}
