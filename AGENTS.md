# AGENTS.md

本项目是 Agent 驱动的轻量级 Go + Vue3 全栈 Web 快速开发框架：后端唯一 **CloudWeGo Hertz**，前端 **Vue3 + TypeScript + Vite + TDesign**。

编码前必须阅读：

- 本文件
- `README.md`
- `.agents/skills/gosilite-dev/SKILL.md`
- 相关 `idl/`、`conf/`、`application/`、`api/` 文件
- Hertz 官方中文文档：`https://cloudwego.cn/zh/docs/hertz/`
- `docs/hertz/api-index.md`

## 目录边界

```text
main.go              # 唯一生产入口（go run .）
api/handler/         # 生成或手写 Handler：绑定、校验、响应
api/router/          # 路由和中间件注册
api/model/           # IDL 生成的请求/响应模型（勿手改）
api/middleware/      # JWT、权限、ctx cache
application/module.go      # Module 插件契约（Init + RegisterRoutes）
application/application.go # Bootstrap / MountRoutes 装配
application/<域>/          # 插件包：module.go + service/ + dal/
application/infra/   # PostgreSQL、Redis、RustFS 等
conf/                # conf/<env>/conf.yaml
idl/                 # Thrift IDL（契约源）
pkg/                 # errorx、response、logger、bcrypt 等
types/               # errno、consts
web/                 # TDesign Vue3 前端
docs/                # patterns / process / hertz / compose spec
```

依赖方向：

```text
main → api/router → api/handler → application/service → application/dal → application/infra
web → Hertz API
```

## Hertz 约定

- 不使用 Gin，不保留旧的标准库业务 HTTP Handler。
- Handler 签名：`func(context.Context, *app.RequestContext)`。
- 路由：`GET/POST/PUT/PATCH/DELETE/Group/Use`。
- 绑定：`BindAndValidate` / `Bind` / `BindQuery` / `BindHeader` / `BindPath` / `BindForm` / `BindJSON` / `BindByContentType`。
- 以 Hertz 官方文档为准，不猜测 Gin 的 `ShouldBind*` 行为。
- 响应只经 `pkg/utils/response`，信封 `{"code","msg","data"}`。
- 入口集中组装配置、infra、Service、路由；业务模块不得调用 `Spin()`。
- 本项目使用 `server.New` + 显式 Recovery（见 `main.go`）；新服务优先 `server.Default`。

## hz 代码生成

```bash
go install github.com/cloudwego/hertz/cmd/hz@latest
hz update -I idl -idl idl/framework/user.thrift -customize_package idl/package.yaml
```

规则：

- Thrift IDL 是公开 API 契约源文件。
- 生成文件放 `api/model`、`api/router`（及 handler 骨架），不散落根目录。
- 不直接修改生成文件；改 IDL 后重新生成。
- 手写 Service/Repository 不放入生成目录。

## 业务分层与插件

- Handler：绑定、校验、调用 Service、响应映射。
- Service：**具体 `Service` 类型 + `NewService`**，`InitService` 只写一次包级门面；**不写 `I*Service` 接口**。
- Module 插件：`Name` / `Init(ctx, deps)` / `RegisterRoutes(h)`；在 `application.builtinModules()` 注册。
- 路由：IDL → hz 生成 `api/router/**`，由模块 `RegisterRoutes` 挂载；`main` 只调 `MountRoutes`。
- Repository：保留接口 + GORM query。
- Infra：数据库、缓存、对象存储、外部服务。
- DTO 与持久化模型分离；字段结构化，不用 JSON blob 代替业务列。

## 会话与安全

- 会话 JSON 使用 `accessToken` / `tokenType` / `expiresIn` / `refreshToken`（OAuth2 风格，camelCase）。
- 禁止新增 `token`/`expire` 旧字段契约。
- Access token = JWT（`Authorization: Bearer`）；Refresh token 存 Redis，支持 rotation/revoke。
- 密码只存 `pkg/bcrypt` 哈希；JWT secret 显式配置且 ≥32 字符。
- 改密后吊销该用户 refresh；登出吊销 refresh + access denylist。
- 管理 API 默认挂 `permission`；列表接口默认分页。

## 流程与技能

- 需求闭环：`docs/process/requirements.md`（Compose Next 对齐）。
- 设计/编码/测试：`docs/patterns/*.md`。
- 开发技能：`.agents/skills/gosilite-dev/SKILL.md`（`.mimocode/skills/` 有副本）。
- Feature spec：`docs/compose/spec/<feature>.md`。

## 命令

```bash
export PATH="$HOME/.local/go/bin:$PATH"   # 本机 Go
GO_ENV=test go run .                      # 入口 main.go，conf/test/conf.yaml
./script/check.sh                         # go test + vet + build
WEB=1 ./script/check.sh                   # 附带 web build
cd web && npm install && npm run build
cd web && npm run dev
docker compose up -d                      # postgres/redis/rustfs
```

完成前必须 `./script/check.sh` 全绿（需要前端交付时再加 `WEB=1`）。
