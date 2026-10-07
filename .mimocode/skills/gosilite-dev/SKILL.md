---
name: gosilite-dev
description: Use when developing, reviewing, or extending web API features with the gosilite Go + Vue3 framework — Hertz Handler/Service/Repository layering, IDL + hz codegen, OAuth2-style session tokens, unified {code,msg,data} envelope, PostgreSQL/Redis/RustFS infra, TDesign Vue3 frontend, and the Compose-aligned requirements loop. Trigger on gosilite, Hertz API, user module, "用 gosilite 开发接口", or adding a business module.
---

# gosilite 开发技能

把 Web 功能需求变成可验收的 Go + Vue3 代码：**IDL 契约 + 分层实现 + 统一响应 + 前后端联调**。

## Important（硬性约定）

1. **分层单向**：`api/handler → application/service → application/dal/repository → application/infra`，禁止跨层；跨域装配只经 `application.Module`。
2. **Service 无接口**：`type Service struct` + `NewService`；`InitService` 只安装包级门面；Repository 保留接口。
3. **错误只用 `errorx`**：`errorx.New(errno.Xxx, errorx.KV("k","v"))` / `errorx.WrapByCode`；Handler 不吞错，统一 `R.Error`。
4. **响应只走 `pkg/utils/response`**：`SuccessWithData` / `Error` / `Unauthorized` / `InvalidParamRequestResponse`；禁止 Handler 手写 JSON。
5. **契约源是 IDL**：`idl/**/*.thrift` 先改，再 `hz update`；不手改 `api/model` 生成文件。
6. **Hertz API 以官方文档为准**：`https://cloudwego.cn/zh/docs/hertz/`；绑定用 `BindAndValidate`，不发明 Gin 风格 API。
7. **会话用 OAuth2 风格字段**：禁用 `token`/`expire` 混用；JSON 字段 camelCase 与前端 `web/src/api/user.ts` 对齐。
8. **模型转换集中 `convert.go`**：dal model ↔ DTO/VO 的字段映射统一写在对应区域包的 `convert.go`，Handler/Service 正文只调用转换函数，禁止内联逐字段映射散落各处（细则见「模型转换 convert.go」）。
9. **验收**：
   ```bash
   ./script/check.sh          # go test + vet + build
   WEB=1 ./script/check.sh    # 需要前端时
   ```

## 何时使用

- 在本仓库新增/修改 HTTP API 业务功能
- 调整用户会话、权限、登录/刷新/登出契约
- 修 bug、补测试、补 schema
- 评审 gosilite 风格的代码
- 为 Agent 驱动开发扩展脚手架、文档或 skill

## 项目事实（禁止凭记忆写旧框架）

| 项 | 当前实际 |
|----|----------|
| 模块路径 | `gosilite` |
| HTTP 框架 | CloudWeGo Hertz（禁用 Gin / 旧标准库业务 Handler） |
| 入口 | 仓库根目录 `main.go`，`go run .` |
| 配置 | `conf/<env>/conf.yaml`，`GO_ENV=test\|dev\|prod` |
| 响应信封 | `{"code":<int32>,"msg":string,"data"?:any}` |
| 错误 | `pkg/errorx` + `types/errno` 数字码，经 `R.Error` 输出 |
| 会话契约 | `SessionTokenResponse`：`accessToken` / `tokenType` / `expiresIn` / `refreshToken`（camelCase） |
| 数据库 | PostgreSQL + GORM + gen（`application/*/dal`） |
| 缓存 | Redis（refresh token / 吊销） |
| 对象存储 | RustFS（S3 兼容，默认 host `19000`） |
| 前端 | `web/` = TDesign Vue3 Starter（Pinia + Axios + vue-router） |
| 生成代码 | `hz` / thriftgo，落在 `api/model`、`api/router` |

## 需求澄清（开发前必做，对齐 docs/process/requirements.md §1）

需求进入编码前先完成澄清。核心：**能推断的不问，只问会改变实现方向的分叉点**。

### 核心原则

1. **先推断，后询问**：能从上下文/代码库/惯例推出的，绝不问用户。
2. **能假设就假设，但必须标注**：有合理默认值的先假设，输出时写「我按 X 处理」。
3. **只问卡住的**：只有「无法推断 + 影响执行」的才问，且带选项。
4. **给出方案而非只给结论**：必要时给 A/B/C 选项 + 代价。
5. **主动指出问题**：需求不合理、有风险、有更简单方案，要主动说。

### 需求解析（内部完成，不输出）

对每个评估维度做三分类：

- 【可推断】能从上下文推出 → 自己补全
- 【可默认】有合理默认值 → 假设 + 标注
- 【必须问】无法推断 + 影响执行 → 加入询问列表

评估维度：①功能目标/输入输出/业务规则/验收标准/优先级；②技术选型/难度/依赖/兼容性；③代码规范/测试/性能/安全/错误处理/可维护性；④改动范围/影响面/交付形式/验证方式/回滚。

### 信息推断（项目落点）

主动读取：项目文件结构、现有代码与配置、`go.mod` / `web/package.json`、现有模块命名与模式、`AGENTS.md` 与 README、需求措辞和上下文。
推断优先级：**项目现有惯例 > 行业最佳实践 > 通用默认值**。本项目惯例见「项目事实」表（Hertz 分层、`{code,msg,data}` 信封、camelCase、`SessionTokenResponse`、`types/errno`）。

### 假设标注

对「可默认」的项，输出时明确标注：**「假设：……（如有偏差请纠正）」**。

### 关键提问

只问同时满足「无法推断 + 不同选择导致完全不同的实现 + 不问就会做错」的：

- 数量 **≤ 3**，按影响大小排序
- 带选项（A/B/C 或 是/否）
- 说明为什么问（影响什么）

### 澄清输出模板

```text
【我的理解】（1–2 句复述需求）
【我的假设】（如有偏差请纠正）
- 技术栈 / 契约：……
- 输入输出：……
【需你确认】（关键问题，带选项，≤3 个）
1. ……？A. …… B. ……
【方案】
- 推荐：……（理由）
- 备选：……（代价：……）
- 风险 / 影响面 / 前置条件：……
```

### 行为准则

- **应该做**：先读代码/上下文再下判断；能推断的不问；假设要标注；问题少而关键且带选项；主动指出不合理处并给替代方案；说清风险、代价、不确定性。
- **禁止做**：把需求确认清单逐条甩给用户；能推断还去问；开放式提问（「你想要什么？」）；不标注假设就默认执行；闷头写代码不做需求评估；对不合理需求直接拒绝而不给替代方案；过度设计（需求只要 A，别做 A+B+C）。

### 特殊场景

- **需求太模糊**：不逐条问。先按最可能的理解给方案 + 标注假设，再问 1–2 个最关键分叉点。
- **需求不合理**：先说明为什么不合理（技术/逻辑/资源），再给替代方案让用户选。
- **需求有冲突**：指出冲突点，给出取舍建议，请用户拍板。
- **信息不足但可假设**：按最合理假设给完整方案，标注「如果 X 不是这样，方案会变成 Y」，让用户低成本纠偏。

## Instructions：标准开发流程（与 Compose Next 对齐）

1. **定契约**：路由、DTO 字段（camelCase）、错误码、表、权限。
2. **写 spec**（功能值得成文时）：`docs/compose/spec/<feature>.md`，含 `[S1] Problem` / `[S2] Design` / `[S3] Out of Scope` / `Tasks`。
3. **改 IDL** → `hz update -I idl -idl idl/<domain>/<file>.thrift -customize_package idl/package.yaml`。
4. **Repo → Service → Handler → Router**，每层可测；model↔DTO/VO 转换写入对应包 `convert.go`。
5. **前端接入**：`web/src/api/*.ts` 类型与 envelope、`store` / 登录态、必要页面。
6. **TDesign MCP 强制**：写/改 `t-*` 组件前先调 `tdesign-mcp-server` 的 `get-component-docs`（选型 `get-component-list`、样式 `get-component-dom`、升级查 `get-component-changelog`）；配置在仓库根 `.mcp.json`。禁止凭记忆猜组件 API。
7. **测试 + 验收四连**，修到全绿。
8. **评审**（需要时）→ spec `status: delivered`。

## 参考速查

### 插件化 Module

- 契约：`application.Module` — `Name` / `Init(ctx, deps)` / `RegisterRoutes(h)`。
- 注册：`application.builtinModules()`；`main` → `Bootstrap` + `MountRoutes`。
- 内置：`application/framework`（user/role/menu 路由）、`application/agent`（draft 路由）。
- 新业务域：写 `application/<域>/module.go` 实现 Module，挂进注册表；Service 用**具体类型**，不要 `I*Service`。

### 新增业务模块骨架

以 `framework/user` 为参考实现，新模块目录：

```text
idl/<domain>/<feature>.thrift
api/model/<domain>/<feature>.go          # hz/thriftgo 生成，勿手改
api/handler/<domain>/<feature>_controller.go
api/handler/<domain>/convert.go          # 该包所有 model↔DTO/VO 转换
api/router/<domain>/<feature>.go         # 生成
api/router/<domain>/middleware.go        # 生成后可改权限挂载
application/<domain>/service/<feature>/  # service.go + service_impl.go
application/<domain>/dal/model/          # gorm gen 模型
application/<domain>/dal/query/          # gorm gen query
application/<domain>/dal/repository/     # 接口 + 实现
types/errno/<feature>.go                 # 错误码注册
web/src/api/<feature>.ts                 # 前端契约
```

组装：Service 在 `application/.../service.go` 的 `InitService` 注入 `db` / cache / storage；Handler 只调 `XxxSVC`。

### 模型转换 convert.go

所有 **dal model ↔ DTO/VO** 的字段映射（`model.SystemDictType` → `syscfg.DictTypeDTO`、SaveRequest → model 等）一律集中在对应区域包的 `convert.go`，不写在 controller/service 函数旁：

- **Handler 包**：`api/handler/<domain>/convert.go` — dal model → 响应 DTO、请求 DTO → dal model。
- **Service 包**（确有转换时）：`application/<domain>/service/<feature>/convert.go`。
- 粒度按**包**一个文件（如 `api/handler/framework/system/convert.go` 收全部 `toXxxDTO`），不按 feature 拆碎文件；包内 feature 很多时可按 feature 命名 `convert_<feature>.go`，但仍归该包统一管理。
- 命名：`toXxxDTO(m *model.Xxx) *pkg.XxxDTO` / `fromXxxRequest(req *pkg.XxxSaveRequest) *model.Xxx`；列表映射提供 `toXxxDTOs([]*model.Xxx) []*pkg.XxxDTO`，避免 Handler 手写循环 append。
- 转换函数只做字段赋值：**不含业务逻辑、不查库、不返回 error**；需要派生计算的放 Service。
- 存量转换已于本次规范落地时全部迁入各包 `convert.go`（system/syscfg/menu/user/agent handler + dept/menu/user service）；此后新增转换直接写入对应 `convert.go`，不落 controller/service 正文。


### 会话与认证（当前实现）

登录成功返回：

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "accessToken": "…",
    "tokenType": "Bearer",
    "expiresIn": 3600,
    "refreshToken": "…"
  }
}
```

- 路由：`POST /user/login`、`POST /user/refresh`、`POST /user/logout`、`POST /user/password`、`GET /user/info`、`GET /user/profile`
- Access token：Hertz JWT，`Authorization: Bearer <accessToken>`
- Refresh token：Redis 存储，支持 rotation / revoke；改密后吊销该用户全部 refresh
- 登出：可选 `refreshToken` + 吊销当前 access token（denylist）
- 前端：`web/src/api/user.ts` + `web/src/store/modules/user.ts`（单飞 refresh）

新增认证相关接口必须同时改：IDL、Handler、`web/src/api`、本技能中的契约表。

### 错误模型

- 业务错误码在 `types/errno` 用 `code.Register` 注册模板消息。
- `errorx.StatusError` → HTTP 200 + `{"code":<业务码>,"msg":...}`（非 2xx 仅用于鉴权失败 401 / 参数 400）。
- 稳定性敏感错误 `code.WithAffectStability(false)`。
- 前端依赖的错误码（如 `ErrUserAuthenticationFailed = 700012006`）**禁止改号**。

### 配置与基础设施

- 配置结构见 `conf/conf.go`：`server` / `jwt` / `database` / `redis` / `storage`。
- 本地依赖：`docker compose up -d` → PostgreSQL 5432、Redis 6379、RustFS 19000/19001。
- JWT secret 必须 ≥32 字符且显式配置。
- 密码只存 `pkg/bcrypt` 哈希。
- 存储兼容 RustFS：HeadBucket 失败允许 CreateBucket，错误码兼容 `NotFound`/`NoSuchBucket`/`404`。

### 测试要求

- Service：可注入仓库/缓存；密码、会话、权限分支要有用例。
- Handler/HTTP：断言 status + `code/msg/data` envelope，不只测 mock 调用。
- 修 bug 先写失败测试再改。
- 详见 `docs/patterns/testing.md`。

### 系统管理（system-rbac）

- 五域：部门树 / 岗位 / 用户 / 菜单 / 角色（含 data_scope 数据权限与 role_menu 分配）。
- IDL：`idl/framework/system.thrift` → `hz update` 生成 `api/model|router/framework/system`；Handler 在 `api/handler/framework/system` 手写业务。
- 模型：`go run ./script/codegen` 生成 `*.gen.go`；迁移见 `script/migrations/0002_system_rbac.sql`。
- 路由由 framework Module `RegisterRoutes` 挂载；`_systemMw` 已挂 `permission.Default()`。
- 前端：`web/src/pages/system/*` + `web/src/api/system/*` + `web/src/router/modules/system.ts`。
- 字典分组（`script/migrations/0008_dict_group.sql`）：`system_dict_group` + `system_dict_type.group_id`（0=未分组；dictType 仍全局唯一，消费端 options/useDict 不带分组）；分组 CRUD 路由 `/system/dict/group/*`，复用 `system:dict:*` 权限码；字典页为左分组列表 + 右类型表格。

## 必读文档

- `AGENTS.md` — 目录边界与硬规则
- `README.md` — 启动与命令
- `docs/patterns/design.md` — 架构与设计模式
- `docs/patterns/coding.md` — 编码规范
- `docs/patterns/testing.md` — 测试规范
- `docs/process/requirements.md` — Compose 需求闭环
- `docs/hertz/api-index.md` — Hertz 速查
- `application/framework/` + `api/handler/framework/user/` — 端到端参考

写 `web/` 时同时加载 `.agents/skills/frontend-dev/SKILL.md`（s-* 通用组件 + TDesign MCP）。

## 检查清单

- [ ] 无 Handler 业务逻辑 / 无 Repository 外泄漏 SQL
- [ ] model↔DTO/VO 转换集中在对应包 `convert.go`，controller/service 正文无内联逐字段映射
- [ ] JSON 契约 camelCase，与 `web/src/api` 一致；无 `token`/`expire` 旧字段混用
- [ ] 响应只经 `R.SuccessWithData` / `R.Error`，信封 `{code,msg,data}`
- [ ] 错误码在 `types/errno` 注册，禁止魔法数字散落
- [ ] IDL 变更已重新生成，生成文件无手改
- [ ] 列表接口分页；管理接口挂 `permission` 中间件
- [ ] 前端 `t-*` 组件 API 已用 TDesign MCP `get-component-docs` 核对（未凭记忆）
- [ ] `go test ./...`、`go vet ./...`、`go build ./...`、`cd web && npm run build` 全绿
