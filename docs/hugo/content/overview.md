---
title: "概览"
weight: 1
---

gosilite 是一个由 Agent 驱动的轻量级 Go + Vue3 全栈 Web 快速开发框架：后端唯一 [CloudWeGo Hertz](https://cloudwego.cn/zh/docs/hertz/)，前端 Vue3 + TypeScript + Vite + TDesign，基础设施（PostgreSQL / Redis / RustFS）一条 `docker compose` 起齐。它面向要快速交付管理后台类系统的开发者（以及代劳编码的 AI Agent）：分层、生成边界、验收脚本都已固定，新模块按套路重复产出即可。第一次接触请先按[快速开始](/docs/quick-start/)跑通，再回本页看整体结构。

## 技术栈

| 层 | 选型 | 说明 |
|---|---|---|
| 后端框架 | CloudWeGo Hertz v0.10.6 | `go.mod` 声明；仓库唯一 HTTP 框架，不混用 Gin 等 |
| 运行时 | Go | `go.mod` 声明 `go 1.26.0`；README 记录本机为 go1.27.x，装在 `$HOME/.local/go/bin`（可能需手动加入 PATH） |
| API 契约 | Thrift IDL（`idl/`）+ hz 生成 | IDL 是 HTTP 契约的唯一源，先改 IDL 再跑 `hz update` 生成路由与 DTO |
| 前端 | Vue3 + TypeScript + Vite + TDesign | 基于 `tdesign-vue-next-starter`；`web/package.json` 要求 Node ≥22.12.0 |
| 数据库 | PostgreSQL 16 | `docker compose` 容器 `postgres:16-alpine` |
| 缓存 | Redis 7 | 容器 `redis:7-alpine`；承载 refresh token 轮换与 denylist（吊销名单） |
| 对象存储 | RustFS（S3 兼容） | 容器映射 19000（S3 API）/ 19001（控制台 UI） |
| 文档站 | Hugo | 源码在 `docs/hugo/`，产物由 Hertz 托管在 `/docs/`；`script/bootstrap-hugo.sh` 可装 pinned 版本到 `script/bin/hugo` |
| 接口文档 | Swagger UI | `GET /swagger/index.html`，由 Handler 上的 swag 注解生成 |
| 服务监听 | `:8888` | `conf/conf.yaml` 的 `server.address`；`GO_ENV` 默认 `test` |

## 架构依赖方向

```text
main.go
  └─ application.Bootstrap / MountRoutes              # 装配：初始化服务 → 挂路由
       └─ Module.RegisterRoutes → api/router → api/handler   # 业务模块插件注册路由
api/handler → application/framework/service → dal → application/infra
web (Vite) ── HTTP ──→ Hertz API (:8888)
```

- **Module 插件**：每个业务域实现 `Init` + `RegisterRoutes`（接口在 `application/module.go`），统一登记在 `application/application.go` 的 `builtinModules()`，入口只做装配。
- **handler → service → dal → infra**：Handler 只做参数绑定、校验与统一回包；业务规则在 Service；SQL 在 dal（`model`/`query` 由 gorm gen 生成，`repository` 手写仓储层）；PG / Redis / RustFS 的连接与驱动在 `application/infra/`。
- **web → Hertz API**：前端开发服务器把 `/api` 前缀剥掉后转发到 `:8888`，`/docs` 原样透传（`web/vite.config.ts`）；浏览器永远只访问同一个源，不直连后端端口。

## 仓库地图

| 路径 | 职责 |
|---|---|
| `main.go` | 唯一生产入口：加载配置、装配、中间件、`/ping`、`/swagger`、`/docs` 托管 |
| `api/handler/` | Handler：绑定、校验、调 Service、经 `pkg/utils/response` 统一回包 |
| `api/router/` | 路由与权限挂载（骨架由 hz 从 IDL 生成，禁止手写 `Register`） |
| `api/model/` | hz/thriftgo 生成的请求响应 DTO（`DO NOT EDIT`） |
| `api/middleware/` | JWT 校验、权限、上下文缓存、操作日志等中间件 |
| `application/` | 后端主体：`application.go` 装配；`base/` 通用支撑；`framework/service/` 业务域服务；`framework/dal/` 表模型与仓储；`infra/` 基础设施驱动 |
| `conf/` | 单一配置 `conf/conf.yaml`；密钥与环境差异走 `.env.<GO_ENV>` / 环境变量（模板 `.env.example`） |
| `idl/` | Thrift IDL，HTTP 契约源（改路由/DTO 先改这里） |
| `pkg/`、`types/` | 通用工具（response、logger、bcrypt 等）与集中错误码 `errno` |
| `script/` | `check.sh` 验收、`migrations/` 幂等迁移 SQL（0002–0022 共 21 个）、`codegen/` 脚手架模板、Hugo 脚本 |
| `web/` | TDesign Vue3 前端（`src/{api,components,pages,store,router,utils}`） |
| `docs/hugo/` | Hugo 文档站源码与产物 `public/` |
| `docs/compose/spec/` | feature spec（需求 / 设计 / 任务） |
| `AGENTS.md` | 给编码者的硬性约定（目录边界、生成边界、验收命令） |

## 能力清单

开箱即用的业务能力（多数在管理端有对应页面，功能细节见[后端](/docs/backend/)与[前端](/docs/frontend/)）：

| 能力 | 说明 |
|---|---|
| RBAC 权限 | RBAC 即"基于角色的访问控制"：用户 → 角色 → 菜单/按钮，权限码形如 `system:user:list`，绑定关系存 `system_role_menu` |
| 用户与会话 | 登录与个人信息、JWT + refresh token 轮换、在线会话页、改密后吊销该用户 refresh token |
| 组织 | 部门（dept）、岗位（post）树与人员归属 |
| 字典与参数 | 数据字典（下拉、标签的集中配置）与全局参数（`system_config` 表，config 服务） |
| 通知公告 | 公告发布、已读记录 |
| 日志 | 操作日志（`operlog` 中间件自动采集）、登录日志与错误日志（对应 `system_oper_log` / `system_login_log` / `system_error_log`） |
| 监控 | 四页：在线会话、数据、缓存、系统接口（Swagger UI）；另有独立 Prometheus `/metrics` 出口（`:9091`，与业务端口隔离） |
| Swagger 接口文档 | 管理端"系统接口"页内嵌 Swagger UI；`swag init` 从注解生成 `doc.json` |
| OSS 对象存储 | 存储配置 / 文件 / 规则三个服务（ossconfig、ossfile、ossrule，对应 `system_oss_config` / `system_oss` / `system_oss_rule`），对接 RustFS 等 S3 兼容存储 |
| 统一文件 | 上传下载收口到 `/common/file/*`，台账由 ossfile 自动记录，业务只存 `fileId` |
| 开发文档站 | 即本站：Hugo 生成静态文件，Hertz 托管在 `/docs/`，管理端"开发文档"页以 iframe 内嵌 |

## 下一步

- 首次跑通项目：[快速开始](/docs/quick-start/)
- 后端分层、逐模块与生成边界：[后端](/docs/backend/)
- 前端路由、组件与开发规范：[前端](/docs/frontend/)
- 踩坑速查：[常见问题](/docs/faq/)
- 命令与配置速查：[参考](/docs/reference/)
