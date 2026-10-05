# 测试模式（Testing Patterns）

目标：Agent 改完业务后能快速确认行为，测试不复制生产逻辑。当前仓库测试基线较薄，**新功能与修 bug 必须按本规范补测**。

## 分层策略

| 层 | 测什么 | 怎么测 |
|----|--------|--------|
| Service | 业务规则、校验、会话轮换、not_found | 真实 Repository 或轻量替身 + 可注入 cache |
| Repository | 查询、分页、错误映射 | 测试库（PostgreSQL testcontainer/本地 docker）或 sqlite 策略（若引入） |
| Handler/HTTP | 状态码、`code/msg/data`、鉴权头 | `httptest` + Hertz Engine 或 handler 直调 |
| 前端 | 请求封装、token 刷新并发 | `npm run build:type` / 后续 vitest（若引入） |
| 组装 | 路由挂载、中间件顺序 | 启动冒烟：`/ping`、`/user/login` |

## 原则

1. **表驱动**：同构输入输出用 `tests := []struct{...}`。
2. **少 mock**：优先真实实现；接口替身只在无法落地（外部 OSS、时钟）时使用。
3. **测公共行为**：断言 HTTP code / 业务 `code` / 持久化结果，不重复实现生产算法。
4. **回归必补**：修 bug 先写能复现的失败测试，再改代码。
5. **禁止测试专用生产 API**；时钟等注入点用 `WithXxx` 且生产默认合理。

## 已有测试

- `application/framework/service/user/session_test.go`：refresh 签发/校验、轮换吊销、access denylist（内存 CacheClient）。

## 会话与认证测试要点

- 登录成功：`data` 含 `accessToken`/`tokenType`/`expiresIn`/`refreshToken`，无 `token`/`expire`。
- 刷新：旧 refresh 轮换后失效；并发只成功一次（与前端单飞逻辑匹配）。
- 登出：refresh 吊销 + access denylist。
- 改密：旧 refresh 全部失效；错误密码映射 `ErrUserInfoInvalidateCode`。
- 未带 `Authorization: Bearer` 访问受保护路由 → 401 envelope。

## 命名约定

- `TestLoginIssuesSessionTokens`、`TestRefreshRotatesRefreshToken`、`TestChangePasswordRevokesSessions`
- 失败分支显式命名意图：`TestLoginInvalidPassword`、`TestRefreshRejectedAfterRevoke`

## 基础设施

- 集成测试依赖：`docker compose up -d`（postgres/redis/rustfs）。
- 不依赖开发机残留数据；测试自建用户/清理。
- JWT secret 用测试专用短配置，不读生产密钥。
- RustFS/S3 测试可只测错误码兼容与 bucket 创建路径，避免绑定大文件。

## 前端验收

```bash
cd web && npm run build        # vue-tsc + vite
```

登录链路手工/自动化冒烟：

1. 错误密码 → 统一错误提示（MessagePlugin/Toast）
2. 正确登录 → 进入布局
3. 401 后自动 refresh 或跳转登录

## 验收命令

在仓库根目录：

```bash
export PATH="$HOME/.local/go/bin:$PATH"   # 本机 Go 安装于 ~/.local/go 时
go test ./...
go vet ./...
go build ./...
cd web && npm run build
```

四条全绿才算可交付。失败先修代码或补测试，不跳过、不 `--skip`。

## MCP / 双通道说明

当前主干是 HTTP + 前端，**没有**业务 MCP 双通道要求。若未来引入 agent 工具暴露，必须复用同一 Service，并在 spec 中单独定义 tool 名与 schema；不要在未写 spec 时顺手加 MCP。
