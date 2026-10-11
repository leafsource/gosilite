---
feature: tool-menu-devdocs
status: delivered
updated: 2026-10-11
branch: develop
commits: f88d456..58ded18
---

# 系统工具菜单 + Hugo/Hertz 开发文档

## Report

**What was built** — 新增顶级菜单「系统工具」（id 49，`/tool`），把「系统接口」4704 从系统监控原地改挂其下（id/权限/角色绑定不变，仅 component 迁 `tool/api/index`、页面名 ToolApi），并新增「开发文档」4901（`tool/docs/index`，iframe 内嵌 `/docs/`）。文档站由 `docs/hugo/` 的 Hugo 工程生成（扁平 content、自写极简 layouts、离线确定性构建），Hertz 在 main.go 用官方推荐的 `adaptor.HertzHandler` 包装标准库 `http.FileServer`（StripPrefix 后托管，不用 `app.FS`）挂在 `/docs`，根目录走 conf `docs.dir` / `GOSILITE_DOCS_DIR`；`/docs` 前缀加入 operlog 跳过名单。构建链：`script/bootstrap-hugo.sh`（pin hugo 0.167.0，linux 下载 tarball / darwin `go install` 源码构建）+ `script/build-docs.sh`（缺 hugo 打印 SKIP 且 exit 0），并挂入 `check.sh`；产物与 script/bin 均 gitignore。Vite 透传 `/docs` 代理到 8888。增补交付（S2.7）：文档站改为 Docsy 式三栏——左栏 sticky 页树（当前页高亮 + h2 锚点子导航）、中栏正文（`render-heading` 注入 `#` 标题锚点）、右栏本页 TOC，窄屏收进页头 `<details>`，纯模板零 JS 零新依赖。

**Verification** — 迁移 0022 在 dev 库连续执行两遍幂等，SELECT 验证 49/4901 落库、4704.parent_id=49、绑定 2 条不变；`WEB=1 ./script/check.sh` exit 0（go test/vet/build + npm build + build docs）；触碰文件 eslint、`vue-tsc --noEmit` exit 0；curl `GET /docs`→301、`/docs/`→200、两子页 200、swagger 200、不存在路径 404；浏览器实测侧栏 系统监控下无系统接口、系统工具下两入口齐全，`/tool/api` 渲染 Swagger、`/tool/docs` 渲染 Hugo 首页、console 无 error；独立 reviewer 结论无 critical。S2.7 增补验证：`build-docs.sh` 无 WARN；三页 DOM 确认左页树当前页高亮 + h2 子导航（4 项）、右 TOC 链接（4 项）、标题锚点（4 处）、首页无 TOC、移动端 `<details>` 内嵌同一 partial；浏览器截图目检三栏布局正常。

**Journey log** —
1. GitHub Release 资产 CDN 在本机仅 ~9KB/s 且 >0.152.2 起无 darwin tar.gz（仅 .pkg 要 sudo）：darwin 改走 `go install` 源码构建（goproxy.cn + checksum DB），linux 保留 tarball 路径。
2. content 若放 `content/docs/` 会产出 `public/docs/*`，StripPrefix 后变成 `/docs/docs/*`：改为扁平 content + layouts 硬编码 `/docs/` 前缀。
3. 曾误用 `curl -sI`（HEAD）验证 301 得到 404——Hertz 只注册了 GET；GET 复测 301 正常。
4. 全仓 eslint 有大量既有失败（monitor/cache、data、online），属基线问题，本次只保证触碰文件干净。
5. 残留风险已记入 Out of Scope：VisibleMenus 不补祖先（自定义角色只绑子菜单会丢菜单）、tabs-router persist 残留旧 /monitor/api 标签 404。
6. Hugo `page.Fragments.Headings` 是**树**（根节点 Level=0，子级在 `.Headings`，文本字段 `.Title` 无 `.Plain`），不是平表——平表在 `HeadingsMap`/`Identifiers`（后者按字典序，不适合导航序）；调试技巧：模板里 `| jsonify` 一发定位字段。

## [S1] Problem

管理端侧栏没有「系统工具」分区；「系统接口」挂在系统监控下语义不合（接口文档不是监控指标）。同时项目缺少一块可持续生长的开发文档区，且文档基建本身要示范「Hugo 生成静态文件 + Hertz 托管」这条集成路线（用户指定：用官方推荐的 adaptor 包装标准库 `http.FileServer`，不用维护状况不佳的 `app.FS`）。

## [S2] Design

### S2.1 菜单迁移 `script/migrations/0022_system_tool.sql`

幂等（`WHERE NOT EXISTS` / 带条件 `UPDATE`），模式仿 0015：

| 动作 | 明细 |
|------|------|
| 新增 M `49` 系统工具 | parent_id=0，path `/tool`，sort=4（系统监控 sort=3 之后），icon `tools`（TDesign MCP search-icon 确认），component `layout`，permission_code `page:root:tool`，perms `''`（M 菜单按 validateMenu 必须清空 perms） |
| 迁移 C `4704` 系统接口 | 仅 `UPDATE`：parent_id 47→49、sort=1、component `monitor/api/index`→`tool/api/index`；**id/perms/permission_code 不变**（`system:apimonitor:list` 与既有 `system_role_menu` 绑定随之保留，无需重建绑定） |
| 新增 C `4901` 开发文档 | parent_id=49，path `docs`，sort=2，component `tool/docs/index`，icon `book-open`，permission_code `page:root:tool:docs`，perms `''`（页面无后端 API；可见性由 role_menu 绑定控制） |
| 角色绑定 | `(49),(4901)` 授予 `system`/`admin` 角色（4704 绑定已存在，不动） |

id 选择：48/4801/4802 已被 0021 回收（历史存储菜单），改用空闲的 **49/4901**，避免复用歧义。

### S2.2 前端

- **移动** `web/src/pages/monitor/api/` → `web/src/pages/tool/api/`，`defineOptions({ name: 'ToolApi' })`——菜单迁移后 fullPath 为 `/tool/api`，后端 `buildFullPath` 拼接、前端 `genRouteName` 生成 `ToolApi`，keep-alive 依赖该名。页面逻辑不变（iframe → `{prefix}/swagger/index.html`）。
- **新增** `web/src/pages/tool/docs/index.vue`，`defineOptions({ name: 'ToolDocs' })`：与系统接口页同构的全高 iframe，`src = '/docs/'`（同源相对路径），load 失败给 t-empty 提示 + 后端直链 + 重试。
- **Vite 代理** `web/vite.config.ts`：proxy 增加 `'/docs' → http://127.0.0.1:8888/`（**不做 rewrite**，路径原样透传；现有 `/api` 条目带前缀剥离，互不影响）。

### S2.3 Hugo 站点 `docs/hugo/`

- 独立 Hugo 工程：`hugo.toml`（`baseURL='/'`、`publishDir='public'`、`disableKinds` 收紧）、自写极简 `layouts/`（baseof/list/single/index），**不引主题、不联网**（构建确定性）。
- 内容（**扁平放根**，不设 section：否则产物路径 `public/docs/*` 在 StripPrefix 后会变成 `/docs/docs/*`）：`content/_index.md`（文档首页/导航卡片）、`content/hugo-hertz.md`（集成原理：Hugo 生成 → Hertz adaptor + http.FileServer 托管 → 菜单入口，含构建/部署说明）、`content/dev-guide.md`（本地开发、check.sh、迁移/菜单/权限约定速查）。站内链接在 layouts 中硬编码 `/docs/` 前缀。
- 产物 `docs/hugo/public/` 进 `.gitignore`（连同 `script/bin/`）。

### S2.4 Hertz 托管（main.go，仿 swagger 挂载）

- conf 新增 `docs.dir`（默认 `docs/hugo/public`，env `GOSILITE_DOCS_DIR` 覆盖，进 `.env.example`）。
- 路由（与 `h.GET("/swagger/*any", adaptor.HertzHandler(...))` 同模式）：
  - `GET /docs` → 301 到 `/docs/`；
  - `GET /docs/*any` → `adaptor.HertzHandler(http.StripPrefix("/docs", http.FileServer(http.Dir(docsDir))))`。
- 目录不存在时 stdlib FileServer 返回 404（可接受，不加特殊兜底）；站内 Hugo 生成 `index.html`，无目录列表风险。
- `/docs` 前缀加入 operlog 跳过名单（`api/middleware/operlog/operlog.go` 的 `/swagger`、`/metrics` 同行），避免静态资源刷操作日志。
- 与 swagger 一致**不挂权限中间件**（公开静态资源；管理端入口由菜单绑定控制）。

### S2.5 构建链

- `script/build-docs.sh`：`hugo` 优先取 PATH，其次 `script/bin/hugo`；**两者都没有 → 打印 SKIP 提示、exit 0**（check 不因缺 hugo 红）；有 hugo 但构建失败 → exit 1。
- `script/bootstrap-hugo.sh`：安装 pinned 版本 Hugo 到 `script/bin/`（一次性引导）。获取策略：linux 下载官方 tar.gz；**darwin 走 `go install github.com/gohugoio/hugo@v0.167.0` 源码构建**（>0.152.2 起官方 release 不再提供 darwin tar.gz、仅 .pkg 需 sudo；且 GitHub 资产 CDN 在本网络仅 ~9KB/s，Go 模块代理 + checksum DB 校验更快更可复现）。
- `script/check.sh`：在 `WEB` 步骤之后追加 `==> build docs (hugo)` 调 `script/build-docs.sh`（任何组合下保持可绿）。

### S2.6 验收

1. dev 库跑 0022 后：侧栏出现「系统工具」> 系统接口 + 开发文档；系统监控下不再有系统接口；两入口按角色可见性正常。
2. `curl -sI http://127.0.0.1:8888/docs/` → 200 text/html（hugo 已构建）；`/docs` → 301；`/swagger/index.html` 不受影响。
3. 浏览器：系统接口、开发文档两页 iframe 正常，keep-alive 缓存名匹配（无 500/白屏）。
4. `WEB=1 ./script/check.sh` 全绿。

### S2.7 文档站 Docsy 式布局（增补，2026-10-11）

用户拍板：不引真 Docsy 主题（需 Hugo Extended + Node 构建链 + 联网依赖，破坏离线约束），在自写 layouts 上仿其**三栏结构**，纯模板零依赖零 JS：

- **布局**：`grid` 三栏——左栏站点页树（sticky、当前页高亮 `aria-current`、当前页下挂 h2 锚点子导航，数据取 `.Site.Home` + `.RegularPages.ByWeight` / `.Fragments.Headings` Level=2）；中栏正文（max-width 阅读宽）；右栏 `.TableOfContents` 本页目录（仅文章页）。
- **标题锚点**：新增 `layouts/_default/_markup/render-heading.html`，渲染时注入 `#` 锚点链接（零 JS）。
- **窄屏**：隐藏左右栏，页头 `<details>` 内嵌同一份页树 partial（`layouts/partials/nav-pages.html`，桌面/移动复用）。
- 内容页 front matter 加 `weight` 固定排序；站内链接仍全部硬编码 `/docs/` 前缀。视觉沿用现有 token（品牌蓝 #0052d9、浅灰侧栏）。

## [S3] Out of Scope

- 真 Docsy/其他 Hugo 主题引入、全文搜索、多语言、版本化文档；TOC scrollspy 滚动高亮（如需可后补十行内联 JS）。
- `/docs` 鉴权/按角色授权（静态公开，与 swagger 同级）；文档内容的持续运营。
- 生产部署形态（容器/CDN）与文档站 CI 自动发布。
- 既有 `api-swagger-page` 行为变更（仅随菜单搬迁，不改其 iframe/权限逻辑）。
- 已知残留（审查记录，交付不修）：`VisibleMenus` 按 role_menu 绑定 id 过滤且不补祖先——自定义角色若只绑 4704 未绑 49 会看不到菜单（本次仅向 system/admin 补绑 49/4901）；前端 tabs-router `persist: true`，升级后浏览器残留的旧 `/monitor/api` 标签点开 404（手动关闭即可）。

## Tasks

- [x] T1: 迁移 0022_system_tool.sql（49 M + 4704 改挂 + 4901 C + 角色绑定）— acceptance: dev 库执行幂等通过；`system_menu` 中 49/4901 存在、4704.parent_id=49 且 component=`tool/api/index`；4704 的 role_menu 绑定数不变 (covers: S2.1)
- [x] T2: Hugo 站点 docs/hugo + build-docs.sh/bootstrap-hugo.sh + gitignore — acceptance: 引导 hugo 后 `script/build-docs.sh` 产出 `docs/hugo/public/index.html`；无 hugo 时脚本 SKIP exit 0；构建产物不入 git (covers: S2.3, S2.5)
- [x] T3: main.go 托管 /docs + conf docs.dir + operlog 跳过 — acceptance: hugo 构建后 `curl /docs/` 200、`/docs` 301、未构建时 404；`go test ./...` 新增/既有用例绿 (covers: S2.4; depends: T2)
- [x] T4: 前端——api 页迁到 tool/、新增 tool/docs 页、vite 代理 /docs — acceptance: eslint+vue-tsc 绿；页面 name 为 ToolApi/ToolDocs 且与 genRouteName(fullPath) 一致 (covers: S2.2; depends: T1)
- [x] T5: 端到端验收——菜单树/双页冒烟 + `WEB=1 ./script/check.sh` — acceptance: S2.6 四条全过 (covers: S2.6; depends: T1, T2, T3, T4)
- [x] T6: Docsy 式三栏布局改版 — acceptance: 构建后三页均渲染左页树/右 TOC/标题锚点，当前页高亮正确，窄屏出 `<details>` 导航；`script/build-docs.sh` 无 WARN (covers: S2.7)
