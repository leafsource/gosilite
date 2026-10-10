# gosilite

由 Agent 驱动的轻量级 Go + Vue3 全栈 Web 快速开发框架。

后端 **CloudWeGo Hertz**，前端 **Vue3 + TypeScript + Vite + TDesign（tdesign-vue-next-starter）**。结构参考相邻 `skyx-cloud-golang`，并收敛为 Agent 可重复生成的开发边界。

## 运行时与 PATH

- 本机 Go：`/Users/leafsource/.local/go/bin/go`（go1.27.x）。若 `go` 不在 PATH：
  ```bash
  export PATH="$HOME/.local/go/bin:$PATH"
  ```
- 入口：仓库根目录 `main.go`（**不是** `cmd/gosilite`）。
- 配置：单一 `conf/conf.yaml`（本地开发默认值），密钥与环境差异经 `.env.<GO_ENV>` 或系统环境变量覆盖（模板见 `.env.example`）；`GO_ENV=test|dev|prod` 选择，默认 `test`（HTTP `:8888`）。生产必须注入 `GOSILITE_JWT_SECRET` / `GOSILITE_DATABASE_DSN` / `GOSILITE_STORAGE_*`。

## 目录结构

```text
main.go                # Hertz 生产入口
api/handler/           # Handler：绑定、校验、响应
api/router/            # 路由与中间件注册
api/model/             # IDL 生成 DTO
api/middleware/        # JWT/权限等中间件
application/           # Service、dal、infra
conf/                  # 环境配置
idl/                   # Thrift IDL（API 契约源）
pkg/                   # errorx、response、logger 等
types/                 # errno、consts
script/codegen/        # 生成模板
web/                   # TDesign Vue3 前端
docs/compose/spec/     # feature spec（Compose Next）
```

依赖方向：`api/handler → application/service → application/dal → application/infra`。  
业务域以 **Module 插件** 挂载：`Bootstrap` Init 服务 → `MountRoutes` 注册路由（见 `application/module.go`）。

## 基础设施（Docker）

```bash
docker compose up -d
```

| 服务 | 端口 | 说明 |
|------|------|------|
| PostgreSQL | 5432 | db `gosilite` / user `gosilite` |
| Redis | 6379 | 会话 refresh / denylist |
| RustFS | 19000 (API), 19001 (UI) | S3 兼容；host 9000 被占用故映射 19000 |

## 启动

```bash
export PATH="$HOME/.local/go/bin:$PATH"
GO_ENV=test go run .          # Hertz :8888
curl -s http://127.0.0.1:8888/ping

cd web && npm install && npm run dev   # Vite :5173
```

登录 API（经 Vite 代理 `/user/*` → `:8888`）：

```text
POST /user/login     → { accessToken, tokenType, expiresIn, refreshToken }
POST /user/refresh
POST /user/logout
POST /user/password
GET  /user/info
GET  /user/profile
```

响应信封：`{"code":200,"msg":"success","data":...}`。会话字段为 OAuth2 风格 camelCase，**不再使用** `token`/`expire`。

## Hertz 与 hz

唯一 HTTP 框架是 Hertz：https://cloudwego.cn/zh/docs/hertz/

```bash
go install github.com/cloudwego/hertz/cmd/hz@latest
hz update -I idl -idl idl/framework/user.thrift -customize_package idl/package.yaml
```

生成代码落在 `api/model`、`api/router`（及 handler 骨架）；契约先改 `idl/`。

## 验收命令

```bash
export PATH="$HOME/.local/go/bin:$PATH"
go test ./...
go vet ./...
go build ./...
cd web && npm run build
```

## Agent 工作规范

1. 编码前阅读 `AGENTS.md`、`.agents/skills/gosilite-dev/SKILL.md`、相关 IDL/Service。
2. 先契约（IDL、DTO、错误码、权限）再实现；流程见 `.agents/skills/gosilite-dev/references/requirements.md`。
3. Handler 只绑定/校验/响应；业务在 Service；SQL 在 Repository。
4. 错误码集中在 `types/errno`；响应只经 `pkg/utils/response`。
5. 修 bug 必须有回归测试；完成必须验收四连全绿。

## 设计目标

- **规范化**：IDL、生成代码、业务、infra 边界清晰。
- **高易用性**：固定模块骨架 + skill 检查清单，Agent 可重复产出。
- **高性能**：Hertz、GORM 连接池、分页与索引、避免重复序列化。
- **高可用**：配置校验、Recovery、RequestID、访问日志、会话吊销、RustFS 兼容容错。
