---
title: "常见问题"
weight: 5
---

以下条目全部来自仓库已知事实（代码 / 脚本 / 迁移 / spec），每条按 **现象 → 原因 → 解法** 三行给出。命令细节见[快速开始](/docs/quick-start/)，约定速查见[参考](/docs/reference/)。

### 1. 接口返回 401，或登录态频繁失效

- **现象**：请求返回 HTTP 401（body 形如 `{"code":401,"msg":"access token has been revoked"}` 等），或用着用着被踢回登录页。
- **原因**：access token 过期 / 被吊销（refresh 轮换、改密会吊销该用户 refresh），`api/middleware/permission/security_auth.go` 直接 `R.Unauthorized` 返回 401。
- **解法**：axios 拦截器已内置——收到 401 只尝试一次 refresh 并重放原请求（`web/src/utils/request/index.ts`），失败才清会话回登录页；若业务调用绕过了 axios（见第 6 条），需自行先刷新 token。

### 2. 多次输错密码后无法登录

- **现象**：即使输入正确密码也提示“用户名或密码错误”。
- **原因**：系统按账号统计失败次数；默认 15 分钟内 5 次失败临时锁定 30 分钟，累计 10 次永久锁定。为避免账号枚举，锁定不会向客户端单独提示。
- **解法**：等待临时锁定结束；永久锁定由管理员在组织架构的用户编辑中提交非空“重置密码”后解除。管理员可在参数管理调整 `sys.login.max_failures`（1–9）、`sys.login.failure_window_minutes`、`sys.login.lock_minutes`。

### 3. 接口提示无权限（常说的 "403"）

- **现象**：接口返回 `{"code":700000007,"msg":"unauthorized access : ..."}`，页面提示无权限（本仓库该场景 HTTP 状态仍为 200，业务码即 700000007）。
- **原因**：`api/middleware/permission` 权限中间件校验失败——当前角色在 `system_role_menu` 没绑对应菜单/按钮，或接口挂载的权限码与菜单 `perms` 不一致。
- **解法**：在角色管理里给角色补绑对应菜单/按钮行；核对 Handler 挂的码与 `system_menu.perms` 一致（模式 `system:<域>:<动作>`，见[参考](/docs/reference/)）。

### 4. 侧栏菜单不显示

- **现象**：库里明明有菜单，管理端侧栏却不出现该目录/页面。
- **原因**：① 角色未绑 `system_role_menu`；② 自定义角色只绑子菜单没绑父级目录——`VisibleMenus` 按绑定 id 过滤且**不补祖先**（spec 已知残留）；③ 菜单 `visible/status` 或父级 `visible` 配置不可见。
- **解法**：补绑缺失的父级目录 id；侧栏唯一数据源是 `GET /menu/routes`，不要去 `web/src/router/modules` 加静态业务路由。

### 5. 菜单点开是 404 页，控制台有 `[router] component not found`

- **现象**：菜单在侧栏可见，点击却落到 404 结果页，控制台 warn `[router] component not found in pages, fallback 404`。
- **原因**：`system_menu.component` 没有指向 `web/src/pages/**` 下的真实文件——要求**相对路径、无 `.vue` 后缀**（如 `monitor/data/index`），大小写必须与文件完全一致（`web/src/utils/route/index.ts` 的 glob 匹配是区分大小写的）。
- **解法**：改菜单 `component` 使其与文件路径逐字符一致（文件不存在就先建页）；改完重新拉 `/menu/routes`。

### 6. 页面不走 keep-alive，来回切换每次都重新加载

- **现象**：Tab 切回旧页面状态丢失、重复请求。
- **原因**：keep-alive `include` 依赖路由 name，而路由 name 由前端 `genRouteName(fullPath)` 生成（`/tool/api` → `ToolApi`）；页面 SFC 的 `defineOptions({ name })` 写错或没写就对不上。
- **解法**：`defineOptions({ name })` 必须等于 `genRouteName(完整 path)`（全大写驼峰）；改 path 后同步改页面名。

### 7. t-upload 上传报 404 / 打到了前端页面路由

- **现象**：用 `t-upload`（`s-upload-image` / `s-upload-file` 同理）上传返回 HTML 或 404，而不是接口 JSON。
- **原因**：`t-upload` 走**原生 XHR，不经过 axios**，不会自动带 `VITE_API_URL_PREFIX`（`/api`）前缀；Vite 代理只转发带 `/api` 的请求并剥前缀给 :8888。
- **解法**：原生 XHR 的 URL 必须自带 `/api` 前缀（如 `/api/common/file/upload`）；另注意原生 XHR 绕过 axios 的 401→refresh，上传前先确保 token 新鲜（组件内已做此处理）。

### 8. 迁移脚本怎么执行

- **现象**：`script/migrations/` 下的 `00XX_*.sql` 不知道在哪儿跑；本机没有 `psql`。
- **原因**：迁移不在程序启动时自动执行，需手动灌库；文件头注释写明「幂等：可重复执行」等前提，部分文件用 `BEGIN` 事务、全表 `public.` 限定。
- **解法**：先看该文件头注释，再执行：

  ```bash
  docker exec -i gosilite-postgres psql -U gosilite -d gosilite < script/migrations/00XX_描述.sql
  ```

### 9. 改了生成物没生效 / 生成物被手改出冲突

- **现象**：改动 `*.gen.go`、`api/model/**` 或 hz 生成的 `Register` 后，重新生成被覆盖，或 review 报「禁手改」。
- **原因**：这些是 `DO NOT EDIT` 生成物——gorm gen 源是数据库表 + `script/codegen`，hz 源是 `idl/**/*.thrift`。
- **解法**：改契约的正确顺序——HTTP/DTO **先改 IDL → `hz update -I idl -idl <file> -customize_package idl/package.yaml` → 再改 Handler**；表结构先写迁移 SQL → `go run ./script/codegen`，永远不直接编辑生成物。

### 10. check 输出 `build docs SKIP`，是不是失败了

- **现象**：`./script/check.sh` 末尾打印 `==> build docs SKIP (hugo 未安装…)`。
- **原因**：`script/build-docs.sh` 在找不到 hugo（PATH 或 `script/bin/hugo`）时打印 SKIP 并 `exit 0`，刻意不阻塞验收。
- **解法**：不是失败；需要构建文档站时先跑 `script/bootstrap-hugo.sh`（安装 pinned 版本到 `script/bin/`），再 `script/build-docs.sh`。

### 11. `WEB=1 ./script/check.sh` 很慢

- **现象**：加了 `WEB=1` 后验收耗时明显变长。
- **原因**：`WEB=1` 会在 go test / vet / build 之外追加 `cd web && npm run build`（完整前端产物构建），随后还要构建文档站。
- **解法**：日常后端改动用 `./script/check.sh` 即可；交付/PR 前再跑 `WEB=1 ./script/check.sh` 做全量验收。

### 12. 全仓 eslint 有一堆既有失败

- **现象**：对整个 `web/` 跑 lint 会看到 monitor/cache、monitor/data、monitor/online 等既有报错。
- **原因**：这些是基线问题（spec 记录为 `PRE-EXISTING`），与本次改动无关，全仓清零不在单次交付范围。
- **解法**：只保证**本次触碰的文件** eslint（及 `vue-tsc --noEmit`）干净；不要顺手大改无关页面。

### 13. 文档链接 `/docs/dev-guide/` 404

- **现象**：旧书签或外部文章里的 `/docs/dev-guide/` 打开是 404。
- **原因**：文档站重构后 `dev-guide` 页已删除，内容拆并到《快速开始》与《参考》两页（见 spec `project-docs` S2.1）。
- **解法**：改链到 [快速开始](/docs/quick-start/)（本地启动 / check）与 [参考](/docs/reference/)（约定速查）。

### 14. 「开发文档」iframe 里点链接套娃出一个中后台（历史问题）

- **现象**：在系统工具 → 开发文档的 iframe 内点站内链接，跳到了前端 SPA 首页（套娃）。
- **原因**：Hugo `baseURL='/'` 时 `.RelPermalink` 生成的链接不带 `/docs` 前缀，被 Vite 的 SPA 兜底接住。
- **解法**：已修复——`docs/hugo/hugo.toml` 设 `baseURL = '/docs/'`，站内链接自动带前缀；自建嵌入页时同理，且 iframe 宽度（700–900px）下要保住左栏断点。
