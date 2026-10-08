---
feature: config-consumption
status: delivered
updated: 2026-10-06
branch: develop
commits: a04f9d3a..working-tree
---

# 参数消费：系统名称与分页默认值接入

## Report

**What was built** — `0006` 将 `sys.app.name` 转 `global` 类型并补「全局」字典行（幂等）；免登录端点 `GET /system/config/global`（手写路由脱离 `rootMw`，AGENTS 允许的跳 hz 方式，`/system/config/key` 保持登录）配 `ListByType/ListGlobal` 仓储与服务；前端 `store/modules/config.ts` 双通道加载（`loadGlobal` 匿名启动拉取并更新 `document.title`，`loadPrivate` 登录后 guard 拉取 `sys.page.size`，失败不置 loaded 可重试），触点覆盖登录页标题、侧栏系统名、`s-table` 分页默认值与六页初始 `pageSize`，`dict.ts` 类型兜底同步补「全局」。

**Verification** — 0006 双跑 exit 0 第二遍 0 行；匿名 `curl /system/config/global` 200 含 `sys.app.name`、匿名 `/key` 401、登录态 `/key` 200；`go build`+vet、`go test ./application/framework/service/config/`（含 `TestListGlobal`）PASS；`WEB=1 ./script/check.sh` PASS；eslint PASS；浏览器验收 5/5（匿名登录页标题 gosilite、改名→新登录页实时变→已恢复 gosilite、分页 20 条/页证明消费生效、侧栏名、console 0 error）；独立评审 0 CRITICAL / 1 MAJOR / 3 MINOR——**MAJOR 已修**（`GlobalList` 错误路径 `R.Failed(200)` 会把失败伪装成成功且泄露 err 原文 → 改 `R.Error`，复验 build+curl+check 全绿）；MINOR 一并处理（`pageDefault` 改 `Number.isInteger` 防 20.5/hex；登录页挂 `loadGlobal` 重试点；余下为文档收口即本节）。

**Journey log** — ①生成 Register 整组继承 `rootMw=permission.Default`，单条豁免登录做不到——免登录路由只能手写脱离组注册（AGENTS 有明文允许）；②`R.Failed(ctx,c,consts.StatusOK,err)` 会产生 `code=200,data:null`，被本仓前端 transform 判成功——匿名端点错误路径必须 `R.Error`；③库里 `sys.page.size` 本就是 20（非种子 10），正好成为「消费生效」的现成铁证（此前硬编码 10）；④SPA 退出登录不重拉 global，标题整页刷新才更新（记录为已知行为）；⑤手写路由与生成 Register 的兄弟静态路径（/key /list）经 radix 无冲突。

## [S1] Problem

参数管理（`system_config`）已交付且种子齐备（`sys.app.name`=gosilite、`sys.page.size`=10），前端 `getConfigByKey` 封装存在但**全站零调用**：

1. 系统名称写死三处——`web/index.html` 静态 title「TDesign Vue Next Starter」、`common.appName`（登录页/侧栏）、`document.title` 从不更新；参数管理里改「系统名称」对任何页面无效；
2. 分页默认大小写死——`s-table` 默认 `pageSize: 10`，user/role/config/dept/post/dict 六页各自 `ref(10)`；改 `sys.page.size` 无效；
3. 登录页要显示系统名，但 `/system/config/key` 是登录后端点（feature 1 故意保留的口需要配套方案）。

**用户拍板（Grill）**：不放开 `/system/config/key` 的登录要求；**新增参数类型「全局（global）`**——将 `sys.app.name` 转为 `global` 类型，通过**新端点 `GET /system/config/global` 免登录获取**。

## [S2] Design

> **Workspace override**：沿用前两个 feature——当前 `develop` 工作区（`a04f9d3a..working-tree`），不开 worktree。

### S2.1 数据与迁移（`script/migrations/0006_config_consumption.sql`，幂等、`public.` 限定）

- `UPDATE public.system_config SET config_type='global' WHERE config_key='sys.app.name' AND config_type IS DISTINCT FROM 'global'`；
- `sys_config_type` 字典补数据行 `全局 → global`（list_order=3，is_default=0）——类型下拉（feature 2 已接字典）自动出现「全局」；
- 前端 `store/modules/dict.ts` 的 `DEFAULT_FALLBACK.sys_config_type` 同步补 `{label:'全局', value:'global'}`（跨 feature 小改，属参数消费的一部分）；
- `sys.page.size` 保持 `system` 类型（登录后经 `/system/config/key` 读取）。

### S2.2 后端：免登录全局端点（~~手写路由~~ → 已迁移 hz 生成）

> **2026-10-08 迁移**：AGENTS 已移除「纯手写系统管理路由可跳过 hz」条款（路由必须 hz 生成）。
> 本端点改为**独立 IDL 文件** `idl/framework/system_config_global.thrift`（namespace
> `framework.syscfgglobal`）→ hz 生成独立 `Register` 与 `middleware.go`，其中 `rootMw`
> 保持 `return nil`（不挂登录）——匿名语义不变；原 `syscfg/global.go` 手写路由与
> `syscfg/config_global.go` 已删除，Handler 迁至 `api/handler/framework/syscfgglobal`。

- ~~**手写路由** `api/router/framework/syscfg/global.go`：`r.GET("/system/config/global", ...)` **不经过** syscfg 包的 `rootMw`（`permission.Default()`），由 `application/framework/module.go` `RegisterRoutes` 挂载。~~
  - ~~**为何不走 hz/IDL**：生成的 Register 整组继承 `rootMw`，无法对单条路由豁免登录；AGENTS 明确允许「纯手写系统管理路由可跳过 hz，直接 handler+router 挂进 Module」。`/system/config/key` 的 IDL 与登录态保持不动。~~（已由上述独立 IDL 方案取代；`/system/config/key` 的 IDL 与登录态始终未动。）
- Handler `GlobalList`（现位于 `api/handler/framework/syscfgglobal`）：调 `SystemConfigSVC.ListGlobal(ctx)` → `{code,msg,data:{list:[ConfigGlobalDTO…]}}`（JSON 字段与原 ConfigDTO 一致）。
- Service `ListGlobal` → Repository `ListByType(ctx, "global")`：`config_type='global' AND status=0 AND del_flag=0`，按 id 升序。
- 语义约束（记录）：`global` 类参数=匿名可读的前端品牌/展示类配置；业务与敏感参数不得放入 global。

### S2.3 前端：config store 与触点

- **`store/modules/config.ts`**：
  - state：`values: Record<string,string>`、`globalLoaded/privateLoaded/pending` 标记；
  - `loadGlobal()`：`request.get('/system/config/global', { withToken:false, skipAuthRefresh:true })`，写入 values；成功后 `document.title = values['sys.app.name'] || document.title`；失败不置 loaded 可重试；
  - `loadPrivate()`：登录后 `getConfigByKey('sys.page.size')` 写入 values（幂等只跑一次）；
  - getters：`appName`（原样返回或 `''`，调用方 `configStore.appName || t('common.appName')` 兜底，store 不耦合 i18n）、`pageDefault`（`Number(values['sys.page.size'])` 合法且 >0 否则 10）。
- **加载时机**：`main.ts` 启动即 `getConfigStore().loadGlobal()`（登录页尽早拿到名字）；路由 guard 在已有 `getUserInfo` 同分支 `await loadPrivate()`（仅登录态、只一次）。
- **触点**：
  1. `web/index.html` 静态 title 保留为 JS 加载前的兜底；
  2. 登录页标题 `login/index.vue` → `configStore.appName || t('common.appName')`；
  3. 侧栏 `SideNav.vue` 版本行同上；
  4. `s-table`：`pageSize` 默认改从 `pageDefault`（props 默认 undefined，`pagination` 计算取 `props.pageSize ?? pageDefault`）；
  5. 六页初始 `pageSize: ref(configStore.pageDefault)`（user/role/config/dept/post/dict 的 type/data 两处）——页面 setup 晚于 guard 的 `loadPrivate`，取到的是参数值。

### S2.4 测试与验收边界

- 迁移 0006 双跑幂等；`curl` **匿名** `GET /system/config/global` 200 且含 `sys.app.name`；**匿名** `GET /system/config/key?key=sys.app.name` 仍 401（登录态 200，回归确认）。
- `./script/check.sh`、`WEB=1 ./script/check.sh` 全绿；eslint 改动文件绿；config service 如有可测分支补 `go test`。
- 浏览器验收（skyx/Test1234）：①未登录打开登录页 → 标题与 `document.title` 显示 `sys.app.name` 当前值；②字典…参数管理页把「系统名称」改为自定义值 → 刷新登录页与浏览器标签随之变化 → 改回；③`sys.page.size` 改 20 → 重新登录进任一列表页，分页每页 20 → 改回 10；④console 无新增 error。
- 独立评审无 CRITICAL。
- **已知边界（评审与浏览器验收记录）**：①SPA 内退出登录不重新拉取 global，系统名/标签需整页刷新更新；②`loadGlobal` 重试=启动一次+进入登录页各一次，再失败留待下次会话；③`pageDefault` 只接受正整数，非法值回落 10。

## [S3] Out of Scope

- 参数变更的跨会话热推送（Redis/广播）；刷新或重登即生效。
- `global` 类型的权限模型扩展（该类=匿名可读，写入仍需 `system:config:add/edit`）。
- 其余业务参数语义化（如站内公告、皮肤等新参数的录入）。
- 字典关联（feature 2 已交付）除 `DEFAULT_FALLBACK.sys_config_type` 补一行外的任何变更。

## Tasks

- [x] T1: 0006 迁移（sys.app.name→global + 字典行 + dict.ts 兜底同步）+ repo/service `ListByType/ListGlobal` — acceptance: 双跑幂等；登录态可查 global 列表含 sys.app.name (covers: S2.1 S2.2)
- [x] T2: 手写 `GET /system/config/global`（handler+router+Module 挂载） — acceptance: 匿名 curl 200 返回 global 参数；匿名 `/system/config/key` 仍 401；登录态 key 200 (covers: S2.2; depends: T1)
- [x] T3: config store（loadGlobal/loadPrivate/appName/pageDefault）+ 触点（main/guard/登录页/侧栏/s-table/六页/dict 兜底） — acceptance: `WEB=1 ./script/check.sh` 绿；登录页与标签显示参数值；列表页初始 pageSize 取自参数 (covers: S2.3; depends: T1)
- [x] T4: 端到端验收（匿名 curl/改名实时/改分页生效/回归 key 401） + 独立评审 + finalize — acceptance: S2.4 步骤逐条通过；评审无 CRITICAL (covers: S2.4; depends: T1 T2 T3)
