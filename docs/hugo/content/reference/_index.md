---
title: "参考"
weight: 6
---

本页是全项目速查：命令、配置、响应信封、权限码与菜单 id、Swagger 注解、目录边界。一切数据与仓库逐行核对（`conf/conf.go`、`.env.example`、`script/schema/v0.1.0.sql`、`AGENTS.md`）；集成原理见子页 [Hugo×Hertz](/docs/reference/hugo-hertz/)，流程性内容见[快速开始](/docs/quick-start/)。

## 一、命令速查

| 场景 | 命令 | 说明 |
|------|------|------|
| 基础设施 | `docker compose up -d` | PG :5432 / Redis :6379 / RustFS :19000-19001（db、user 均 `gosilite`） |
| Go PATH | `export PATH="$HOME/.local/go/bin:$PATH"` | 本机 go1.27 不在 PATH 时先导出 |
| 后端启动 | `GO_ENV=test go run .` | Hertz :8888；探活 `curl -s http://127.0.0.1:8888/ping` |
| 前端启动 | `cd web && npm install && npm run dev` | Vite :3002（`web/vite.config.ts`）；代理 `/api` 剥前缀转发 :8888，`/docs` 原样透传 |
| 登录 | `skyx` / `Admin@123` | dev/test 库系统管理员 |
| 验收 | `./script/check.sh` | go test + vet + build + 文档站（缺 hugo 时 SKIP） |
| 全量验收 | `WEB=1 ./script/check.sh` | 追加 `cd web && npm run build` |
| hz 生成 | `hz update -I idl -idl idl/framework/user.thrift -customize_package idl/package.yaml` | 源：`idl/**/*.thrift`；产物 `api/model`、`api/router` 等（勿手改） |
| 表模型/CRUD 生成 | `go run ./script/codegen` | 源：数据库表；DSN 可用 `GOSILITE_CODEGEN_DSN` 覆盖 |
| 初始化 v0.1.0 schema | `./script/init-schema.sh` | 仅空库；已有核心表时拒绝执行，不做删除或覆盖 |
| Swagger 再生成 | `swag init -g main.go -o ./docs/swagger --parseDependency --parseInternal` | 改 Swag 注解后执行；UI `GET /swagger/index.html` |
| 文档站引导 | `script/bootstrap-hugo.sh` | 一次性装 pinned hugo（0.167.0）到 `script/bin/` |
| 文档站构建 | `script/build-docs.sh` | 产物 `docs/hugo/public`；缺 hugo 打印 SKIP 且 exit 0 |

## 二、配置表（conf.yaml 键 → 环境变量覆盖）

加载顺序：启动时读 `conf/conf.yaml`，再加载仓库根 `.env.<GO_ENV>`（默认 `.env.test`），**同名系统环境变量优先**；空值不覆盖。`GO_ENV=test|dev|prod` 决定加载哪个文件（默认 `test`）。下表逐行对照 `conf/conf.go` 的 `applyEnvOverrides` 与 `.env.example`：只有列出变量的键可被环境变量覆盖，其余键一律以 yaml 为准。

| conf.yaml 键 | 本地默认值 | 环境变量（可覆盖） |
|--------------|------------|--------------------|
| `server.node` | `1` | — |
| `server.address` | `:8888` | `GOSILITE_SERVER_ADDRESS` |
| `server.enable_pprof` | `true`（仅非 prod 生效） | — |
| `server.enable_gzip` | `true` | — |
| `server.enable_access_log` | `true` | `GOSILITE_SERVER_ACCESS_LOG`（`true`/`false`） |
| `server.log_level` | `info` | — |
| `server.log_file_name` | `log/skyx-cloud.log` | — |
| `server.log_max_size` | `10` | — |
| `server.log_max_backups` | `50` | — |
| `server.log_max_age` | `3` | — |
| `jwt.issuer` | `skyx-cloud` | — |
| `jwt.secret` | 本地开发默认值 | `GOSILITE_JWT_SECRET`（≥32 字符；**prod 必填**） |
| `database.dsn` | `host=127.0.0.1 user=gosilite password=gosilite dbname=gosilite port=5432 ...` | `GOSILITE_DATABASE_DSN`（**prod 必填**） |
| `database.max_idle_conns` | `10` | — |
| `database.max_open_conns` | `100` | — |
| `database.conn_max_lifetime` | `3600` | — |
| `database.conn_max_idle_time` | `600` | — |
| `redis.address` | `127.0.0.1:6379` | — |
| `redis.password` | 空 | `GOSILITE_REDIS_PASSWORD` |
| `storage.access_key` | `gosilite` | `GOSILITE_STORAGE_ACCESS_KEY`（**prod 必填**） |
| `storage.secret_key` | `gosilite-secret` | `GOSILITE_STORAGE_SECRET_KEY`（**prod 必填**） |
| `storage.endpoint` | `http://127.0.0.1:19000` | `GOSILITE_STORAGE_ENDPOINT` |
| `storage.region` | 空 | — |
| `storage.bucket_name` | `skyx` | — |
| `storage.domain` | `127.0.0.1:19000` | `GOSILITE_STORAGE_DOMAIN` |
| `monitor.enable_metrics` | `true` | — |
| `monitor.metrics_addr` | `:9091` | — |
| `monitor.metrics_path` | `/metrics` | — |
| `monitor.sql_slow_ms` | `200` | — |
| `monitor.sql_max_templates` | `200` | — |
| `monitor.firewall_enable` | `true` | — |
| `monitor.gorm_postgres_metrics` | `false` | — |
| `docs.dir` | `docs/hugo/public` | `GOSILITE_DOCS_DIR` |
| `scheduler.log_retention_days` | `30` | `GOSILITE_SCHEDULER_LOG_RETENTION_DAYS` |

要点：

- `GO_ENV=prod` 启动时强制要求环境注入 `GOSILITE_JWT_SECRET` / `GOSILITE_DATABASE_DSN` / `GOSILITE_STORAGE_ACCESS_KEY` / `GOSILITE_STORAGE_SECRET_KEY`，缺任一个直接 panic（`conf/conf.go`）；pprof 在 prod 不注册。
- 另有与 yaml 无关的变量：`GOSILITE_CODEGEN_DSN`（仅 `script/codegen` 用）；前端 `web/.env` 的 `VITE_API_URL_PREFIX=/api`、`VITE_BASE_URL=/`。

## 三、响应信封与会话字段

**信封**（所有接口统一，`pkg/utils/response`）：

```json
{ "code": 200, "msg": "success", "data": {} }
```

| 规则 | 事实 |
|------|------|
| 成功 | HTTP 200 + `code:200` |
| 业务错误 | 通常仍是 **HTTP 200** + 业务 `code`（经 `R.Error` 输出）；参数错误 HTTP 400、认证失败 HTTP 401、兜底异常 HTTP 500 |
| 常见业务码 | `700012006` 认证失败（前端依赖，勿改）· `700000007` 无权限 · `700000004` 会话失效 · `700000006` 参数非法 · `700000003` 账号或密码错误（`types/errno/user.go`） |
| id 序列化 | int64 id 经 `jsonid` 输出为字符串（`response.writeJSON` → `TransformOut`） |

**会话字段**（OAuth2 风格 **camelCase**，禁止 `token`/`expire`）：

```json
{ "accessToken": "...", "tokenType": "Bearer", "expiresIn": 7200, "refreshToken": "..." }
```

| 端点 | 说明 |
|------|------|
| `POST /user/login` | 登录，返回上述四字段 |
| `POST /user/refresh` | 用 refreshToken 换新 accessToken（Redis 存 refresh，轮换/吊销） |
| `POST /user/logout` | 登出（携带 refreshToken） |
| `POST /user/password` | 改密（吊销该用户 refresh） |
| `GET /user/info` | 用户信息，含 `permissions`（供 `v-permission` 按钮隐藏） |

## 四、权限码模式与菜单 id 速查

**模式**：`system:<域>:<动作>`，如 `system:user:list` / `system:role:add` / `system:notice:remove`（动作以 IDL 与迁移实际写法为准）。行类型与码的约定：

| menu_type | 含义 | perms 约定 |
|-----------|------|-----------|
| `M` | 目录 | **必须为空**（validateMenu 校验）；permission_code 用 `page:root:*` |
| `C` | 页面 | `system:<域>:list`，与 permission_code 同码 |
| `F` | 按钮 | `system:<域>:add/edit/remove/...`，挂对应页面下 |

绑定关系存 `system_role_menu`；接口校验在 `api/middleware/permission`；超管豁免判定为角色 `system` 或 `admin`（`IsSuperRole`）。

**菜单 id 速查**（均核实自 `script/schema/v0.1.0.sql`；仅列**当前存在**的菜单）：

| id | 名称 | 父级 |
|----|------|------|
| 1 / 2 | 仪表盘（目录）/ 概览 | 0 / 1 |
| 3 | 系统管理（目录） | 0 |
| 10 | 菜单管理 | 3 |
| 41 / 42 | 岗位管理 / 角色管理 | 3（0002） |
| 43 | 参数管理 | 3（0004） |
| 44 | 组织架构 | 3（0007，吸收原用户/部门管理） |
| 45 | 通知公告 | 3（0011） |
| 46 / 4601 / 4602 | 日志管理 / 操作日志 / 登录日志 | 3 / 46 / 46（0013） |
| 47 / 4701 / 4702 / 4703 | 系统监控 / 在线用户 / 数据监控 / 缓存监控 | 0 / 47 / 47 / 47（0015–0017） |
| **4704** | 系统接口（Swagger 页） | **49** |
| **49** | 系统工具（目录） | 0 |
| **4901** | 开发文档（iframe `/docs/`） | 49 |
| 60 | 对象存储（目录） | 3（0021） |
| 61 / 62 / 63 | OSS 配置 / 文件管理 / OSS 处理规则 | 60（0021） |

已删除、**勿再引用**的 id：12 用户管理、40 部门管理（0007 合并进 44 组织架构）；48 / 4801 / 4802 存储管理（0021 吸收进 62 文件管理）。

## 五、Swag 注解最小集（所有 HTTP Handler 强制）

```go
// Datasource 数据源监控
// @Summary 数据源监控
// @Description 数据源监控接口
// @Tags 监控
// @Accept json
// @Produce json
// @Param data body monitor.DatasourceDTO true "请求参数"   // GET 用 query；无参省略
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/datasource [GET]
func Datasource(ctx context.Context, c *app.RequestContext) { ... }
```

| 注解 | 要求 |
|------|------|
| 必填 | `@Summary` `@Tags` `@Produce` `@Success` `@Router`（大写 Router），缺一即不合格 |
| `@Tags` | 限用：用户 / 系统管理 / 系统配置 / 监控 / 日志 / 通知 / 菜单 |
| `@Param` | 有请求才写；POST/PUT/DELETE 用 `body api/model.结构体`，GET 用 `query` |
| `@Success` | 统一 `{object} map[string]interface{}`（信封在 `response` 包） |
| `@Router` | 路径与 IDL `api.get/post/...` 一致；改路由先改 IDL → `hz update` → 再补注解 |
| 再生成 | `swag init -g main.go -o ./docs/swagger --parseDependency --parseInternal` |
| 挂载 | `main.go` 空导入 `_ "gosilite/docs/swagger"`；UI `GET /swagger/index.html`；管理端「系统接口」菜单内嵌 |

## 六、目录速查

| 路径 | 职责 | 边界 |
|------|------|------|
| `main.go` | 唯一生产入口（`server.New` + Recovery，挂 `/docs`、`/swagger`） | 手写 |
| `api/handler/` | 绑定、校验、调 Service、回包（必带 Swag 注解） | 手写为主 |
| `api/router/` | 路由注册 | hz 生成骨架；**禁止手写 `Register`** |
| `api/model/` | 请求/响应 DTO | hz/thriftgo 生成，勿手改 |
| `api/middleware/` | JWT、权限（`permission`）、操作日志（`operlog`）、ctx cache | 手写 |
| `application/` | `application.go` 装配 + 各域 `module.go` / `service/` / `dal/`（model、query、repository） | 装配手写；`*.gen.go` 由 gorm gen 生成 |
| `conf/` | `conf.yaml` 单一配置；密钥走 `.env.<GO_ENV>` 与环境变量 | 手写 |
| `idl/` | Thrift 契约源（hz 的输入） | 手写 |
| `pkg/`、`types/` | errorx、response、errno | 手写 |
| `script/schema/v0.1.0.sql` | v0.1.0 最终 DDL、索引与系统种子 | 手写；仅空库初始化；后续版本策略单独决定 |
| `script/codegen/` | gorm/service/repository 脚手架模板与生成器 | 手写模板 |
| `web/src/` | Vue3 前端（api / components / pages / store / router / utils） | 手写 |
| `docs/compose/spec/` | feature spec（Compose Next） | 手写 |
| `docs/hugo/` | 文档站工程（本页来源），产物 `public/` 不入 git | 手写 |

依赖方向：`main → application → Module.RegisterRoutes`；`handler → service → dal → infra`；`web → Hertz API`。

---

下一步：[Hugo×Hertz 集成原理](/docs/reference/hugo-hertz/) · 上一页：[常见问题](/docs/faq/) · 返回[首页](/docs/)
