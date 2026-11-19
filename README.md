# botgo

QQ 机器人（QQ 群 / 单聊）官方协议 Go SDK。

fork 自 [tencent-connect/botgo v0.2.1](https://github.com/tencent-connect/botgo)——上游已停止维护，
本 fork 持续跟进官方 2026 年起新增的群 / 单聊开放接口，QQ 频道部分与上游保持一致。

## 安装

```bash
go get github.com/by-yitong/botgo
```

## 快速上手

```go
import botgo "github.com/by-yitong/botgo"

api := botgo.NewOpenAPI(appID, tokenSource).WithTimeout(60 * time.Second)
info, _ := api.GetGroupInfo(ctx, groupOpenID)
```

WebSocket / Webhook 事件接入、token 管理见 `event/`、`token/`、`sessions/` 各包。

## 相对上游的主要增补（官方 2026-08 起）

- 群管理：群信息 / 机器人群内状态 / 禁言 / 成员列表与详情 / 黑名单 / 入群审批
- 富媒体：群与单聊分片直传（upload_prepare / upload_part_finish）
- 素材管理（media-images，markdown 图片平台转存）
- 图文卡片发送（msg_type=8）、@全体成员内嵌标签
- 用户基本信息（GET /v2/users/{user_openid}/info）
- 流式发送单聊消息（POST /v2/users/{user_openid}/stream_messages）
- 指令面板管理（详情 / 修改 / 关联对象）
- 生成分享链接（POST /v2/generate_url_link）
- 入群自动审批策略管理（创建 / 修改 / 执行 / 白名单 / 删除）
- 事件：群成员加入（GROUP_MEMBER_ADD）、入群申请（GROUP_JOIN_REQUEST）等

## License

Apache-2.0（继承上游）
