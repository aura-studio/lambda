# SQS 模块需求与约定（当前实现）

本文档汇总当前仓库中 `sqs` 模块的协议与路由行为，作为与调用方/上游系统对接的固定契约。

## 目标

- 在 AWS Lambda 的 `SQSEvent` 触发下，按“类似 HTTP”的路由系统处理请求：`InstallHandlers / Use / HandleAllMethods / NoRoute / NoMethod`。
- 支持常见路径：`/`、`/health-check`、`/api/*path`、`/_/api/*path`、`/wapi/*path`、`/_/wapi/*path`。
- 严格要求对外 API 访问必须以 `/api/`（或 `/wapi/`）开头；不存在对无前缀路径的“TrimPrefix 容忍”。
- 协议层不再依赖 SQS `MessageId` 做业务相关性。
- 可选响应：需要响应时产出 OutgoingMessage（由上层发送回 SQS）。不需要响应时允许 `ClientSqsId` 为空。

## 消息协议（Request / Response）

- SQS Message `Body`：**base64(protobuf 二进制)**
- protobuf schema：见 [sqs/sqs.proto](sqs/sqs.proto)

### Request

字段顺序（proto field number）：

1. `request_sqs_id`：客户端标识（需要响应时必填）
2. `response_sqs_id`：服务端/响应队列标识（非空表示“需要响应”）
3. `correlation_id`：可选相关性字段（实现会透传到 Response）
4. `path`：路由路径（如 `/api/pkg/version/route`）
5. `payload`：请求内容（bytes）；在引擎内当作字符串传给 handler（`string(payload)`）

### Response

字段顺序（proto field number）：

1. `request_sqs_id`
2. `response_sqs_id`
3. `correlation_id`
4. `payload`：响应内容（bytes）；当前实现写入 `[]byte(responseString)`
5. `error`：保留字段（当前实现未填充；失败时通常不产出响应而是 batch failure）

## 路由系统

实现位于 [sqs/router.go](sqs/router.go) 与 [sqs/handlers.go](sqs/handlers.go)。

- `InstallHandlers()` 默认注册：
    - `/` → `OK`
    - `/health-check` → `OK`
    - `/api/*path` → `API`
    - `/_/api/*path` → `Debug` + `API`
    - `/wapi/*path` → `WAPI`
    - `/_/wapi/*path` → `Debug` + `WAPI`
    - `NoRoute` → `PageNotFound`
    - `NoMethod` → `MethodNotAllowed`（当前路由不区分 method，主要用于对齐 http 形态）

### 通配符行为（`*path`）

- 对于 pattern `/api/*path`：
    - 输入 `/api/pkg/version/route`
    - `Context.ParamPath` 为 `/pkg/version/route`（带前导 `/`）
- 这使得内部 invoker 看到的路径与 dynamic 的“真实路径”对齐。

## Engine 行为

实现位于 [sqs/engine.go](sqs/engine.go)。

- `Invoke` 逻辑：
    - 若 `BatchMode` 为开启状态 → 调用 `HandleSQSMessagesWithResponse`（支持部分重试）。
    - 否则 → 调用 `HandleSQSMessagesWithoutResponse`（失败时重试整个批次）。

- 每条 SQS record：
    1. base64 decode + proto unmarshal 得到 `Request`
    2. 创建 `Context{RawPath: request.Path, Path: request.Path, Request: string(request.Payload)}`
    3. router `dispatch`
    4. 若 `Context.Err != nil`：
        - 若 `SuspendMode` 为开启状态 → 立即停止处理并返回错误（中断批次）
        - 否则 → 记为 batch failure，继续处理下一条
    5. 若 `request.ResponseSqsId != ""` 且 `ReplyMode` 为开启状态 → 需要响应：
        - 要求 `request.RequestSqsId != ""`，否则记为 batch failure
        - 构造 `Response{RequestSqsId, ResponseSqsId, CorrelationId, Payload}`
        - 输出 `OutgoingMessage{QueueID: request.ResponseSqsId, Body: base64(proto(Response))}`
    6. 若 `request.ResponseSqsId == ""` → 不需要响应（允许 `RequestSqsId` 为空）

### Partial batch failure

- 返回 `events.SQSEventResponse.BatchItemFailures`，仅标记失败 record 的 `MessageId`。

## 测试覆盖

回归测试位于 [tests/sqs_handler_test.go](tests/sqs_handler_test.go)，覆盖：

- 无效 body → partial failure
- `/api/*path` 的前缀剥离（通配符参数）
- `/pkg/...`（无 `/api/` 前缀）→ NoRoute → partial failure
- `ServerSqsId` 非空才产出响应；无响应时 `ClientSqsId` 可为空
