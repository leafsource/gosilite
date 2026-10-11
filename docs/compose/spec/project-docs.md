---
feature: project-docs
status: delivered
updated: 2026-10-11
branch: develop
commits: 78cc5b2..420daa1
---

# 项目文档（Hugo 站多页 · 面向新用户）

## Report

**What was built** — 文档站新增一套面向新用户的完整项目文档，根级精简为 6 页：《概览》（技术栈/架构/仓库地图/能力清单）、《快速开始》（依赖→后端→前端→登录→验收→文档站）、《后端》（分层与 Module 约定 + 13 个 service 域逐模块 + 生成边界与迁移）、《前端》（目录/动态路由/API 层/18 组件表/11 页面域/开发规范）、《常见问题》（13 条 现象→原因→解法）、《参考》（命令速查、33 键配置表逐行对齐 conf.go 与 .env.example、信封与会话字段、菜单 id 速查含已删除 id 注记、Swag 最小集），外加迁入 `reference/` 的《Hugo × Hertz 集成》。原《开发指南》按约定拆并删除，首页卡片改指概览/快速开始。全部陈述以仓库代码/迁移/脚本为事实源，与 README、AGENTS 冲突处一律以代码为准并在文中标注。

**Verification** — `script/build-docs.sh` 无 WARN；站内链接 169 条死链 0；事实抽查 11 项（账号/端口/env 变量/service 与组件清单/菜单 id/hertz 版本/信封字段/vite 代理等）全过；`WEB=1 ./script/check.sh` exit 0；独立 reviewer 30+ 处抽查报 1 critical（菜单 id 表混入已删除 id 12/40/48 系）+ 3 non-critical，全部修复后复审 4/4 PASS、无新 critical。

**Journey log** —
1. 4 个并行 subagent 按 spec 大纲分页撰写、事实源点名，效率与一致性兼得；统一由主控构建/死链检查，避免并发写 public 冲突。
2. 三处仓库内过时表述被文档纠正（README Vite 5173→以 vite.config 3002 为准；AGENTS component 缺失 500→代码实为 404；spec 笔误 config_global→实为 system_config）——文档以代码为准是「无臆测」的核心。
3. Hugo section 页 footer 用 `.File.ModTime` 直接崩溃（section 的 File 类型无该字段），改为「Lastmod 年份 <2 则只显站名」。
4. 菜单 id 归源必须区分**基线 INSERT**（script/postgres.sql）与**迁移 UPDATE/DELETE**（0002 起），否则会把已删除 id 写进「当前存在」表——reviewer critical 的根因。
5. dev-guide 的「约定速查」2 行落《前端》而非全在《参考》，属跨页合理归位（non-critical 记录在案）。

## [S1] Problem

文档站目前只有首页 + Hugo×Hertz 集成 + 开发指南 3 页，缺少一份**面向新用户的完整项目文档**：项目是什么、怎么跑起来、前后端各模块做什么、踩坑与速查。需要基于仓库事实（无臆测）、语言简洁、先结论后细节地从头到尾写清。

## [S2] Design

### S2.1 信息架构（根级精简，共 6 个根节点）

```text
content/
  _index.md                    # 首页（保留，导航卡片改指新页）
  overview.md        weight=1  # 概览
  quick-start.md     weight=2  # 快速开始
  backend.md         weight=3  # 后端
  frontend.md        weight=4  # 前端
  faq.md             weight=5  # 常见问题
  reference/
    _index.md        weight=6  # 参考（索引）
    hugo-hertz.md              # 由根级迁入（原 content/hugo-hertz.md）
```

- **删除** `content/dev-guide.md`：其「本地启动/check」并入《快速开始》，「约定速查表」并入《参考》；`hugo-hertz.md` 内指向 `/docs/dev-guide/` 的链接改为 `/docs/quick-start/`、`/docs/reference/`。
- 首页 `_index.md` 两张卡片改为：概览 + 快速开始；正文提及的链接同步更新。
- 树排序：首页 → 概览 → 快速开始 → 后端 → 前端 → 常见问题 → 参考（含子页 hugo-hertz）。

### S2.2 撰写原则

1. **无臆测**：一切陈述以仓库代码/README/AGENTS/迁移为准；不确定的不写。命令必须与 `script/`、`README.md` 实际一致（含 `PATH` 导出、`GO_ENV=test`、测试账号 `skyx`/`Admin@123`）。
2. **先结论后细节**：每页开头 2–4 句结论段，再展开；每模块先一句话职责，再要点。
3. **面向新用户**：不假设已读 AGENTS；术语首次出现给一句解释；跨页引用用站内链接（`/docs/...`）。
4. 篇幅克制：单页正文目标 ≤300 行；表格优先于长段落。

### S2.3 各页内容大纲（正文的事实来源在括号内）

- **概览**：结论（gosilite 是什么、谁用它）→ 技术栈表（Hertz v0.10.6 / Vue3+TS+Vite+TDesign / PG / Redis / RustFS / Hugo 文档站）→ 架构依赖方向图（`main → application → Module.RegisterRoutes`；`handler → service → dal → infra`；`web → Hertz API`）→ 仓库地图（AGENTS.md 一节目录表）→ 能力清单（RBAC、组织、字典/参数、通知、日志、监控四页、Swagger、OSS 存储、统一文件、开发文档站）。
- **快速开始**：前置（Docker、Go 1.27 PATH、Node； hugo 可选）→ `docker compose up -d`（端口表：PG 5432 / Redis 6379 / RustFS 19000-19001）→ 后端 `GO_ENV=test go run .`（:8888，`/ping` 探活）→ 前端 `cd web && npm install && npm run dev`（vite 端口 3002，代理规则 `/api` 剥前缀、`/docs` 透传）→ 登录（`skyx` / `Admin@123`，会话字段 camelCase）→ 验收 `./script/check.sh`（`WEB=1` 含前端）→ 文档站 `script/bootstrap-hugo.sh` + `script/build-docs.sh`。
- **后端**：先行约定（分层职责表、Module 插件机制、响应信封、错误码 `types/errno`、Swagger 注解强制）→ **逐模块**（每模块：职责一句话 + 关键表/路由/权限码 + 入口文件）：认证与会话（JWT+refresh 轮换，`api/middleware/permission`、`application/base/ctxutil`）、用户与 RBAC（user/role/menu，`system:*:list` 模式、`system_role_menu`）、组织（dept/post）、字典与参数（dict/config_global）、通知公告（notice）、日志（syslog：操作/登录，operlog 中间件）、监控（monitor service：在线/数据/缓存/接口 Swagger 四页 + `/metrics`）、存储（ossconfig/ossfile/ossrule + `common.file` 统一上传下载）、infra（orm 连接池、cache Redis、storage RustFS S3、imagex）→ 生成边界表（idl→hz、表→gorm gen、templates→codegen；`DO NOT EDIT` 规则）→ 迁移（`script/migrations/` 幂等、当前 21+ 个）。
- **前端**：技术栈与目录（`web/src/{api,components,pages,store,router,utils}`）→ 动态路由（唯一来源 `GET /menu/routes`、`genRouteName(fullPath)` 与 `defineOptions` 对齐、keep-alive）→ 登录与 API 层（axios 信封、`VITE_API_URL_PREFIX=/api`、401 刷新）→ 通用组件清单（`s-table`、`s-form-dialog/drawer`、`s-upload-image/file`、`s-select-dict` 等 15 个，命名规则 `S<Name>`）→ 页面域（system / monitor / user / notice / tool / dashboard）→ 开发规范（t-* 必须 TDesign MCP、组件三抽、弹窗表单 >5 项用 drawer）。
- **常见问题**（每条：现象 → 原因 → 解法，全部来自仓库已知事实）：接口 401/刷新、403 无权限、侧栏菜单不显示（role_menu/component 路径/大小写）、原生 XHR 忘带 `/api`（t-upload 案例）、迁移怎么执行（docker exec psql）、生成物被手改（`*.gen.go`/`api/model`/`Register`）、旧链接 `/docs/dev-guide/` 404、iframe 打开开发文档套娃（历史，已由 `baseURL=/docs/` 修复——说明机制）、`WEB=1 check` 慢、全仓 eslint 有既有失败（`PRE-EXISTING`，只保证触碰文件干净）、hugo 未装时 SKIP。
- **参考**：`_index.md` = 索引 + **命令速查表**（启动/验收/生成/迁移/文档站）+ **配置表**（conf.yaml 键 → env 覆盖变量，对照 `.env.example` 与 `conf/conf.go`）+ **信封与会话字段** + **权限码模式**（`system:<域>:<动作>`、菜单 id 速查：3 系统管理 / 47 系统监控 / 49 系统工具 / 4704 / 4901，来源 `script/migrations`）+ **Swag 注解最小集** + 目录速查；子页 `hugo-hertz.md` 原文迁入。

### S2.4 验收

1. 构建：`script/build-docs.sh` 无 WARN；根级树恰 6 个节点（+ 首页），`reference/` 下 1 子页。
2. 死链：构建产物中所有站内 `href="/docs/..."` 目标文件均存在（脚本核对）。
3. 事实抽查：登录账号、端口、命令、配置变量、菜单 id、组件清单与仓库逐项一致（抽查 ≥10 处）。
4. `WEB=1 ./script/check.sh` 全绿（确认未破坏既有交付）。
5. 独立 reviewer 按 S2.2 原则核查「无臆测」。

## [S3] Out of Scope

- API 逐接口参考（Swagger UI 已覆盖，不重复罗列）。
- 英文版、版本化文档、全文搜索。
- README.md 重写（仅保持其与新文档不冲突，不改 README）。
- 文档站主题/布局改动（沿用现有 Docsy 式 tree + TOC）。

## Tasks

- [x] T1: 信息架构落地——6 页骨架 + weight、dev-guide 拆并删除、hugo-hertz 迁入 reference/、首页卡片与互链更新 — acceptance: build 后根级树 6 节点 + reference 1 子页；站内 href 全部可解析；无 `/docs/dev-guide/` 残留 (covers: S2.1)
- [x] T2: 撰写《概览》《快速开始》 — acceptance: 大纲要点全覆盖；命令/端口/账号与 README、conf.yaml 实测一致 (covers: S2.2, S2.3; depends: T1)
- [x] T3: 撰写《后端》（逐模块） — acceptance: 13 个 service 域与 application/framework/service 目录一一对应，权限码/表名来自迁移与 IDL (covers: S2.2, S2.3; depends: T1)
- [x] T4: 撰写《前端》 — acceptance: 组件清单与 web/src/components 一致，路由/代理规则与代码一致 (covers: S2.2, S2.3; depends: T1)
- [x] T5: 撰写《常见问题》《参考》 — acceptance: 配置变量表与 conf.go/.env.example 逐行一致；菜单 id 与迁移一致 (covers: S2.2, S2.3; depends: T1)
- [x] T6: 验收——死链脚本、事实抽查 ≥10 处、`WEB=1 ./script/check.sh`、独立 reviewer 无臆测核查 — acceptance: S2.4 五条全过 (covers: S2.4; depends: T2, T3, T4, T5)
