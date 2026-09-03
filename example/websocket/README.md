# WebSocket Lambda 示例与部署指南

本目录包含一个最小可用的 WebSocket Lambda 示例：

- `main.go` — 基于本库 `websocket` 包的示例程序（echo tunnel + ping 路由 + 连接钩子）
- `tester.html` — 浏览器直连的 WebSocket 测试工具（无需安装任何东西）
- 本文档 — API Gateway + Lambda 的完整部署步骤

## 架构

```
浏览器/wscat ──wss──► API Gateway (WebSocket API) ──事件──► Lambda (main.go)
                            │                                  │
                            │  $connect / $disconnect /        │  PostToConnection
                            │  $default / ping                 │  (@connections 管理 API)
                            ◄──────────── 推送 ─────────────────┘
```

连接由 API Gateway 托管，Lambda 只在事件发生时被调用（建连、断连、每条消息），
回推一律走 `@connections` 管理 API，而不是 Lambda 的返回值。

## 消息协议

| 客户端发送 | 命中路由 | 行为 |
|---|---|---|
| `{"path": "/echo/v1/echo", "payload": {...}}` | `$default` | echo tunnel 原样回显 payload |
| `{"path": "/echo/v1/time"}` | `$default` | 返回当前 UTC 时间 |
| `{"action": "ping"}` | `ping` | 轻量心跳，回 `{"pong": <毫秒时间戳>}` |

回推格式为 `{"path": ..., "payload": ...}`；业务错误为 `{"path": ..., "error": ...}`。

建连时可带 query 参数做鉴权演示：`wss://.../dev?token=demo`，`token=deny` 会被
`$connect` 拒绝（HTTP 403，连接不建立）。

## 部署步骤（AWS CLI）

以下假设 `REGION=ap-southeast-1`，按需替换；`<ACCOUNT_ID>` 用
`aws sts get-caller-identity --query Account --output text` 获取。

### 1. 构建打包

`provided.al2023` 运行时要求可执行文件名为 `bootstrap`：

```bash
cd example/websocket
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .
zip function.zip bootstrap
```

### 2. 创建 Lambda 函数

需要一个执行角色（已有可复用；没有则建一个挂
`AWSLambdaBasicExecutionRole` 的策略即可）：

```bash
aws lambda create-function \
  --function-name ws-example \
  --runtime provided.al2023 \
  --architectures arm64 \
  --handler bootstrap \
  --role arn:aws:iam::<ACCOUNT_ID>:role/<执行角色名> \
  --zip-file fileb://function.zip \
  --region $REGION
```

### 3. 创建 WebSocket API

```bash
# 创建 API，路由选择表达式决定消息按 body 里的 action 字段分流
aws apigatewayv2 create-api \
  --name ws-example \
  --protocol-type WEBSOCKET \
  --route-selection-expression '$request.body.action' \
  --region $REGION
# → 记下 ApiId（下称 <API_ID>）

# 创建 Lambda 代理集成
aws apigatewayv2 create-integration \
  --api-id <API_ID> \
  --integration-type AWS_PROXY \
  --integration-uri arn:aws:lambda:$REGION:<ACCOUNT_ID>:function:ws-example \
  --payload-format-version 1.0 \
  --region $REGION
# → 记下 IntegrationId（下称 <INTEGRATION_ID>）

# 四条路由都指向同一个集成，引擎内部按 RouteKey 再分发
for KEY in '$connect' '$disconnect' '$default' ping; do
  aws apigatewayv2 create-route \
    --api-id <API_ID> \
    --route-key "$KEY" \
    --target integrations/<INTEGRATION_ID> \
    --region $REGION
done

# 允许 API Gateway 调用 Lambda
aws lambda add-permission \
  --function-name ws-example \
  --statement-id apigw-ws-invoke \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:$REGION:<ACCOUNT_ID>:<API_ID>/*/*" \
  --region $REGION

# 创建自动部署的 stage
aws apigatewayv2 create-stage \
  --api-id <API_ID> \
  --stage-name dev \
  --auto-deploy \
  --region $REGION
```

### 4. 授权 Lambda 回推消息（关键，漏了这步推送全部 403）

给 Lambda 执行角色加内联策略：

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "execute-api:ManageConnections",
    "Resource": "arn:aws:execute-api:<REGION>:<ACCOUNT_ID>:<API_ID>/dev/*"
  }]
}
```

### 5. 验证

连接地址：`wss://<API_ID>.execute-api.$REGION.amazonaws.com/dev`

```bash
# 方式一：wscat
npx wscat -c "wss://<API_ID>.execute-api.$REGION.amazonaws.com/dev?token=demo"
> {"path": "/echo/v1/echo", "payload": {"hello": "world"}}
< {"path":"/echo/v1/echo","payload":{"echo":{"hello":"world"},"route":"/echo"}}
> {"action": "ping"}
< {"pong":1712345678901}

# 方式二：浏览器打开 tester.html，填入地址即可
```

## 运行限制（API Gateway 侧，不可调）

- 单连接最长 2 小时，空闲 10 分钟自动断开——客户端需要重连逻辑
- 单条消息最大 128KB
- 每个事件 Lambda 须在 29 秒内返回
- 鉴权只在 `$connect` 执行一次，之后的消息网关不再校验

## 成本要点

- 计费 = 连接分钟（$0.25/百万分钟）+ 消息条数（$1/百万条，双向都计）
- Lambda 按调用次数 + 时长另计，每次事件就是一次调用
- 浏览器无法发送协议层 ping 帧，应用层心跳建议分钟级间隔（空闲超时是 10 分钟），
  秒级心跳会让消息费成为账单大头
- 协议层 ping/pong 控制帧由网关处理，不产生消息费和 Lambda 调用

## 排错

| 现象 | 排查 |
|---|---|
| 连接直接失败 | 看 Lambda CloudWatch 日志 `/aws/lambda/ws-example`；`$connect` 返回非 200 即拒绝（示例里 `token=deny` 会触发） |
| 能连上但收不到回包 | 九成是第 4 步 `execute-api:ManageConnections` 没配，日志里会有 403 |
| 消息无响应且无日志 | 路由选择表达式没命中：body 里带 `action` 字段但没有对应 route；不带 `action` 的走 `$default` |
| 连接 10 分钟掉线 | 空闲超时，需要心跳或重连 |
| 502 | Lambda 报错或超时，看 CloudWatch |

## 生产化清单（示例未包含）

- `$connect` 换真实鉴权（示例只做演示），建议加 Lambda authorizer
- 连接映射落 DynamoDB（`connectionId ↔ userId`，带 TTL），推送时按 userId 寻址
- 推送失败遇 410 Gone 要顺手删连接记录（库里的 `IsGone`）
- 广播场景用 SNS/EventBridge 扇出，不要单函数循环推
- 自定义域名 + access log + 按路由的 throttling
