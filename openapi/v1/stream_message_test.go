package v1

// 本地 fork 补充的流式发送单聊消息测试（官方 2026-11 文档新增）：
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

func TestPostStreamC2CMessageFirstChunk(t *testing.T) {
	var gotPath, gotMethod string
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotPath = r.URL.Path
		gotMethod = r.Method
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"id":"SM1","timestamp":"2026-11-19T10:00:00+08:00","ext_info":{"ref_idx":"7"},"remain_msg_len":900}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	rsp, err := api.PostStreamC2CMessage(context.Background(), "U1", &dto.StreamMessageReq{
		InputMode:   "append",
		InputState:  1,
		ContentType: "text",
		ContentRaw:  "你好，正在生成",
		MsgID:       "MSG1",
		MsgSeq:      1,
	})
	if err != nil {
		t.Fatalf("PostStreamC2CMessage: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/v2/users/U1/stream_messages") {
		t.Errorf("path=%q want /v2/users/U1/stream_messages", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method=%q want POST", gotMethod)
	}
	var req struct {
		InputMode   string `json:"input_mode"`
		InputState  int32  `json:"input_state"`
		ContentType string `json:"content_type"`
		ContentRaw  string `json:"content_raw"`
		MsgID       string `json:"msg_id"`
		StreamMsgID string `json:"stream_msg_id"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if req.InputMode != "append" || req.InputState != 1 || req.ContentType != "text" ||
		req.ContentRaw != "你好，正在生成" || req.MsgID != "MSG1" {
		t.Errorf("首片请求体字段错: %+v", req)
	}
	if req.StreamMsgID != "" {
		t.Errorf("首片不应带 stream_msg_id: %q", req.StreamMsgID)
	}
	want := dto.StreamMessageRsp{
		ID:           "SM1",
		Timestamp:    "2026-11-19T10:00:00+08:00",
		ExtInfo:      &dto.MessageExtInfo{RefIdx: "7"},
		RemainMsgLen: 900,
	}
	if rsp.ID != want.ID || rsp.Timestamp != want.Timestamp || rsp.RemainMsgLen != want.RemainMsgLen {
		t.Errorf("响应字段错: %+v", rsp)
	}
	if rsp.ExtInfo == nil || rsp.ExtInfo.RefIdx != "7" {
		t.Errorf("ext_info.ref_idx 应反序列化为 7: %+v", rsp.ExtInfo)
	}
}

func TestPostStreamC2CMessageNextChunk(t *testing.T) {
	var bodyBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		bodyBytes, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{"id":"SM1","timestamp":"2026-11-19T10:00:01+08:00"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if _, err := api.PostStreamC2CMessage(context.Background(), "U1", &dto.StreamMessageReq{
		InputMode:   "append",
		InputState:  10,
		Index:       1,
		ContentType: "text",
		ContentRaw:  "，已生成完毕",
		StreamMsgID: "SM1",
	}); err != nil {
		t.Fatalf("续片请求: %v", err)
	}
	var req struct {
		InputState  int32  `json:"input_state"`
		Index       int32  `json:"index"`
		StreamMsgID string `json:"stream_msg_id"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("请求体应为合法 JSON: %v body=%q", err, bodyBytes)
	}
	if req.StreamMsgID != "SM1" || req.Index != 1 || req.InputState != 10 {
		t.Errorf("续片请求体应带 stream_msg_id/index/input_state: %+v", req)
	}
}

func TestPostStreamC2CMessageHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"code":40007,"message":"已下发前缀不可修改"}`)
	}))
	t.Cleanup(srv.Close)

	api := testAPI(t, srv)
	if _, err := api.PostStreamC2CMessage(context.Background(), "U1", &dto.StreamMessageReq{}); err == nil {
		t.Fatal("HTTP 400（40007 前缀不可修改）应报错")
	}
}
