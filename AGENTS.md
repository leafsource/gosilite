# AGENTS.md

本项目是 Agent 驱动的轻量级 Go + Vue3 全栈 Web 快速开发框架：后端唯一 **CloudWeGo Hertz**，前端 **Vue3 + TypeScript + Vite + TDesign**。

编码前必须阅读：

- 本文件
- `README.md`
- `.agents/skills/gosilite-dev/SKILL.md`（后端/全栈）
- `.agents/skills/frontend-dev/SKILL.md`（写 `web/` 时）
- 相关 `idl/`、`conf/`、`application/`、`api/` 文件
- Hertz 官方中文文档：`https://cloudwego.cn/zh/docs/hertz/`
- `docs/hertz/api-index.md`

## Skills 自动使用

满足条件时**无需用户点名**，直接加载对应 skill 并按其执行：

| 条件 | Skill |
|------|--------|
| 改 Go 业务/API/迁移/会话/分层 | `gosilite-dev` |
| 新增或修改 `web/**/*.vue`、`components/**`、使用 `t-*`、组件重复 ≥3 处、组件命名 | `frontend-dev` |
| 走 Compose Next 多阶段交付 | 已加载的 `compose-next` 流程 |

未显式 `/<skill>` 时，上表条件即视为自动触发；一次任务可同时命中多个 skill。

## 前端组件（摘要，细则见 frontend-dev skill）

1. **通用组件** ≥3 处复用 → `web/src/components/s-<name>/`，组件名 `S<Name>`（现成：`s-table`）。
2. **业务组件** → 模块下 `components/`（如 `pages/system/dict/components/`），不进全局。
3. 同一 UI **第 3 次**拷贝 → 停止复制，抽 `s-*`。
4. `t-*` API 必须先 **TDesign MCP** `get-component-docs`，禁止凭记忆。

## 前端 TDesign MCP（强制）

配置见仓库根 `.mcp.json`：

```json
{
  "mcpServers": {
    "tdesign-mcp-server": {
      "command": "npx",
      "args": ["-y", "tdesign-mcp-server@latest"]
    }
  }
}
```

Tools：`get-component-docs`（写代码前必查）、`get-component-list`（选型）、`get-component-dom`（DOM/样式）、`get-component-changelog`（升级前）。

规则：

1. 新增或修改 `t-*` 前，先 `get-component-docs` 核对 props/events。
2. 组件报错或属性不生效时，先 `get-component-dom` / docs，再改代码。
3. 升级 `tdesign-vue-next` 前必查 changelog。
4. 选型不确定用 `get-component-list`。
5. 前端完成后 `WEB=1 ./script/check.sh`。

---

## 一、目录边界

```text
main.go                     # 唯一生产入口（go run .）【手写】
api/handler/                # Handler：绑定、校验、调 Service、回包【以手写为主】
api/router/                 # 路由注册【hz 生成骨架】
api/model/                  # 请求/响应模型【hz/thriftgo 生成，勿手改】
api/middleware/             # JWT、权限、ctx cache【手写】
application/module.go       # Module 插件契约【手写】
application/application.go  # Bootstrap / MountRoutes / builtinModules【手写】
application/<域>/module.go  # 域插件 Init + RegisterRoutes【手写】
application/<域>/service/   # 业务服务【手写】
application/<域>/dal/model/ # *.gen.go 由 gorm gen 生成；*.go 手写模型可并存
application/<域>/dal/query/ # gorm gen 查询 API【生成】
application/<域>/dal/repository/ # 仓储接口+实现【手写】
application/infra/          # PG / Redis / RustFS【手写】
conf/                       # conf/<env>/conf.yaml【手写】
idl/                        # Thrift 契约源【手写，是 hz 的输入】
script/codegen/             # gorm/service/repository 脚手架模板【手写模板】
pkg/ types/                 # errorx、response、errno【手写】
web/                        # TDesign Vue3 前端【手写】
docs/                       # patterns / process / hertz / compose spec【手写】
```

依赖方向：

```text
main → application (Bootstrap/MountRoutes)
         └─ Module.RegisterRoutes → api/router → api/handler
api/handler → application/*/service → application/*/dal → application/infra
web → Hertz API
```

---

## 二、生成边界（必须分清）

**总原则：标 `DO NOT EDIT` / `*.gen.go` 的文件禁止手改；只改生成源后重新生成。**

### 1. hz / thriftgo —— 源：`idl/**/*.thrift`

| 输入 | 命令 | 生成物（勿手改） |
|------|------|------------------|
| `idl/**/*.thrift` + `idl/package.yaml` | `hz update -I idl -idl <file> -customize_package idl/package.yaml` | `api/model/**`、`api/router/**/xxx.go`（`Register`）、部分 handler 骨架 |

- `api/router/**/middleware.go` 生成后**允许改**权限挂载（`permission.Default()` 等），但保持函数骨架。
- `api/router/register.go` 的 `GeneratedRegister` 仅供 hz 追加；**运行时由 Module 的 `RegisterRoutes` 挂载**，不要在 `main` 再调 `GeneratedRegister`。
- 改路由/DTO：**先改 IDL → 再 hz update → 再改 Handler 实现**。

### 2. gorm gen —— 源：数据库表 + `script/codegen/codegen.go`

| 生成物 | 说明 |
|--------|------|
| `application/**/dal/model/*.gen.go` | 表模型，`DO NOT EDIT` |
| `application/**/dal/query/*.gen.go` | 链式查询 API，`DO NOT EDIT` |

- 新表：先加 SQL（`script/migrations/`）或改库 → 再跑 gen/`script/codegen`。
- 手写模型可放在同包非 `.gen.go` 文件（如 `system_rbac.go`），与 gen 共存。

### 3. script/codegen 模板 —— 源：`script/codegen/templates/*.tmpl`

| 模板 | 生成物 |
|------|--------|
| `repository.go.tmpl` | `dal/repository/*_repository.go` 初稿（存在则跳过） |
| `service.go.tmpl` | `service/<域>/service.go` 初稿（**具体 Service，无 I\*Service**） |

- 模板已对齐 `gosilite` 模块路径；不要恢复 `skyx-cloud` 或 `service_impl` 双文件生成。
- 生成后**归手写维护**；再次生成只覆盖不存在的文件。

### 4. 完全手写（永不生成）

- `main.go`、`application/module.go`、`application/application.go`、各域 `module.go`
- Handler 业务逻辑、Service、Repository 方法体、权限中间件、`types/errno`、`conf`、`docs`、`web`、`pkg`
- 迁移 SQL、README/AGENTS/skill

### 边界速查

```text
要改什么          → 动哪里              → 怎么再生
HTTP 契约/路由    idl/*.thrift         hz update
表结构            migrations SQL       gorm gen / codegen
CRUD 样板         templates/*.tmpl     script/codegen（仅新文件）
业务规则          service/*.go         不生成，直接改
装配/插件         application/*.go     不生成
前端              web/src              不生成
```

---

## 三、Hertz 约定

- 不使用 Gin；不用旧标准库业务 HTTP Handler。
- Handler 签名：`func(context.Context, *app.RequestContext)`。
- 绑定：`BindAndValidate` / `Bind` / `BindQuery` / `BindHeader` / `BindPath` / `BindForm` / `BindJSON` / `BindByContentType`。
- 以 Hertz 官方文档为准，不猜测 Gin 的 `ShouldBind*`。
- 响应只经 `pkg/utils/response`，信封 `{"code","msg","data"}`。
- 入口集中组装；业务模块不得调用 `Spin()`。
- 本项目 `main.go` 使用 `server.New` + 显式 Recovery。

---

## 四、业务分层与插件

- Handler：绑定、校验、调 Service、`R.SuccessWithData` / `R.Error`。
- Service：**具体 `Service` + `NewService`**；`InitService` 只安装包级门面；**禁止 `I*Service`**。
- Repository：接口 + GORM（gen query 或 `*gorm.DB`）；不进 Handler。
- Module：`Name` / `Init(ctx, deps)` / `RegisterRoutes(h)`；注册于 `application.builtinModules()`。
- DTO 与持久化模型分离；业务字段结构化，不用 JSON blob。

---

## 五、标准开发流程

```text
1. 定契约     DTO 字段、路由、错误码、表、权限码（system:*:xxx）
2. （可选）   docs/compose/spec/<feature>.md  —— 功能值得成文时
3. 迁移       script/migrations/NNNN_*.sql 幂等 DDL + 种子
4. 生成       表模型 → gorm gen；HTTP → 改 idl 后 hz update
   （纯手写系统管理路由可跳过 hz，直接 handler+router 挂进 Module）
5. 仓储       dal/repository（手写或 codegen 初稿）
6. 服务       service/**（具体类型 + NewService）
7. 装配       域 module.Init / RegisterRoutes；需要时改 builtinModules
8. 前端       web/src/api + pages，envelope 对齐 {code,msg,data}
9. 验收       ./script/check.sh   （WEB=1 可带前端 build）
10. 文档      更新本文件/skill 中与该变更相关的约定
```

顺序约束：

- **先 IDL 后生成，先生成源后手改生成物（禁止直接改生成物）。**
- 先表后 gen，先 Repository 后 Service 后 Handler。
- 改已交付契约必须同步：IDL、错误码、前端类型、skill 契约表。

---

## 六、会话与安全

- 会话 JSON：`accessToken` / `tokenType` / `expiresIn` / `refreshToken`（camelCase）；禁止 `token`/`expire`。
- Access = JWT Bearer；Refresh 在 Redis，rotation/revoke；改密吊销该用户 refresh。
- 密码仅 `pkg/bcrypt`；JWT secret ≥32 字符。
- 管理 API 默认 `permission` 中间件；列表默认分页。

---

## 七、流程与技能文档

- 需求闭环：`docs/process/requirements.md`
- 设计/编码/测试：`docs/patterns/*.md`
- 技能：`.agents/skills/gosilite-dev/SKILL.md`（`.mimocode/skills/` 同步副本）
- Feature spec：`docs/compose/spec/<feature>.md`

---

## 八、命令

```bash
export PATH="$HOME/.local/go/bin:$PATH"
GO_ENV=test go run .
./script/check.sh            # go test + vet + build
WEB=1 ./script/check.sh
cd web && npm run build
cd web && npm run dev
docker compose up -d
```

完成前必须 `./script/check.sh` 全绿（前端交付时加 `WEB=1`）。

**禁止**：`git reset --clean/checkout` 破坏未提交重建现场；引入 Gin/MinIO；手改 `*.gen.go` / `api/model` / hz 生成的 `Register`。
