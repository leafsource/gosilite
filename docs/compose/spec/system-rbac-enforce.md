---
feature: system-rbac-enforce
status: delivered
updated: 2026-10-06
branch: develop
commits: a04f9d3a..working-tree
---

# RBAC 断链修复：菜单过滤 / perms 鉴权 / 按钮级控制

## Report

**What was built** — 三层权限闭环：后端 `/menu/routes` 按 `user_role→role_menu` 过滤（超管 `admin|system` 全量，`ctxutil.IsSuperRole` 统一豁免）并挂登录鉴权；`/system/*` 与 syscfg 全部 per-route 挂 `permission.HPermissions`（dept/tree、config/key、dict/data/options 保持 login-only）；前端 user store 保存 `permissions`，侧栏改为「静态首页 + 后端过滤菜单」唯一业务源（静态业务模块退出侧栏但路由仍可直达），新增 `v-permission` 指令接入七个 system 页面共 29 处按钮；迁移 `0004_rbac_enforce_seed.sql` 建字典/参数菜单与 12 个按钮 perms 行、修正 0002 的 perms 回填 bug、绑定 zc1001→`system_default` 并授予仅列表的 6 菜单子集（防提权）。

**Verification** — `WEB=1 ./script/check.sh` PASS；menu/permission 单测 4+5 用例 PASS；0004 三连执行 exit 0、第二遍 0 行（幂等）、库内数据核验 PASS；curl 验收 8/8（未登录 401、zc1001 子集菜单/user list 200/menu list 信封拒绝 700000007、login-only 200、skyx 全量 12 路由）；浏览器验收 5/5（双角色侧栏差异、按钮隐藏/显示、直达页+越权拒绝、console 0 error）；eslint/stylelint PASS；独立评审 0 CRITICAL（2 MAJOR：均为 spec 文字收口，已在本报告/设计段修正）。

**Journey log** — ①根因不止后端：侧栏「系统管理」来自静态 `system.ts` 进 `fixedRouterList`，后端菜单被 path 去重后根本进不了侧栏——只过滤接口不改侧栏等于没修；②0002 先把 `permission_code(page:*)` 拷进 `perms` 再按 `perms=''` 补真码，条件已不成立，id12/10 的 perms 一直是 page:*，0004 强制修正；③超管实为角色 `system`（dump 无 `admin` 角色），豁免判定收敛为 `IsSuperRole(admin|system)`；④zC1001 子集只授列表不授按钮——否则 `user:add`/`role:assign` 可自提权，同时正好用于验收按钮隐藏；⑤容器内 psql 对裸表名 `system_user` 解析异常（限定 `public.` 后正常），迁移 SQL 全部 schema 限定规避。

## [S1] Problem

`system-rbac` spec 已 delivered，但权限链路三层同时断链，普通用户 zc1001 登录后仍看到全部菜单和数据：

1. 后端 `/menu/routes` 是 `TODO`，直接 `AllMenu` 返回全量，且 menu 路由组中间件全为 `nil`——**未登录也能拿到全量菜单**；
2. `/system/*` 路由组只挂 `permission.Default()`（仅登录），全部 per-route 中间件返回 `nil`——**没有任何 perms 校验**，无权限调用不会 403；
3. 前端 `user` store 丢弃 GetInfo 返回的 `permissions`，`permission` store 拿到菜单全量直接渲染；按钮级 `v-permission` 指令（spec 承诺）整体缺失；
4. 种子数据：zc1001（user id 50）无任何角色行，普通角色 `system_default` 无 `system_role_menu` 记录——即使过滤生效也无数据可验收。

用户纠正的设计点：**前端不做二次路由过滤**——后端过滤后的 `/menu/routes` 是唯一真相，前端直接渲染；前端保存 `permissions` 仅服务于按钮级指令。

## [S2] Design

> **Workspace override**：本 feature 在当前 checkout（`develop` 工作区）实施，不开 `.worktrees/`——既有 RBAC/字典/参数等已交付代码全部以未提交状态存在于该工作区（沿用前两个 spec 的 `a04f9d3a..working-tree` 惯例）。

本 feature 是三段式「关联打通」的第一段（后续：字典关联 `dict-integration`、参数消费 `config-consumption`）。

### S2.1 后端：/menu/routes 按角色过滤 + 登录鉴权

- `api/handler/framework/menu/menu_controller.go` `GetRoutes`：从 ctx 取 `LoginUser` → 按 `user_role → system_role_menu → system_menu` 过滤；**超管（`admin|system`，`ctxutil.IsSuperRole`）返回全量**；无角色返回空列表；剔除按钮行（F）与已删行。
- 路由组挂 `permission.Default()`（`api/router/framework/menu/middleware.go` rootMw），未登录 401。
- **记录偏离**：`system-rbac` spec S2.3 承诺的 `GET /system/menu/routes` 不补建；在既有 `/menu/routes` 原位修复，前端无感。

### S2.2 后端：/system/* 逐路由挂 perms 校验

- 组级保持 `permission.Default()`（登录）；`api/router/framework/system/middleware.go` 与 `api/router/framework/syscfg/middleware.go` 全部 per-route Mw 从 `nil` 改为 `permission.HPermissions("<perms码>")`。
- perms 码以 `system_menu.perms` 种子为唯一依据（`0002` 菜单种子 + `0004` 新建的字典/参数菜单）；动词约定 `list/add/edit/remove`，角色分配用 `system:role:assign`。
- **login-only 例外**（不挂 perms，仅登录）：`GET /system/dept/tree`（用户表单/数据范围选树依赖）、`GET /system/config/key`（运行时参数读取，供 `config-consumption`）、`GET /system/dict/data/options`（下拉选项，供 `dict-integration`）。
- **回显端点（评审后修正记录）**：`GET /system/role/{id}/menus|depts` 挂 `system:role:list`、`GET /system/user/{id}/roles` 挂 `system:user:list`——回显与调用页面的列表权限同权，比 login-only 更严格；浏览器验收确认无功能回归。
- **超管豁免**：`role_key=admin` **或内置 `system`** 不受 perms/菜单过滤限制（库中真实超管角色为 `system`；判定收敛为 `ctxutil.IsSuperRole`，菜单过滤与 perms 中间件共用）。
- 无权限登录用户 → `errno.ErrUserPermissionCode`（信封 700000007，403 语义；HTTP 状态保持 200 为项目既有 response 封装行为）；未登录 → 认证失败。

### S2.3 前端：保存 permissions + 侧栏以后端菜单为唯一来源 + v-permission

- `store/modules/user.ts`：`getUserInfo` 保存 `res.permissions`（不再丢弃），新增 `permissions` getter。
- **侧栏重构**（关键发现：当前侧栏的「系统管理」来自静态 `router/modules/system.ts` 进 `fixedRouterList`，后端菜单被 path 去重后进不了侧栏，因此只过滤后端接口不改侧栏）：
  - `buildAsyncRoutes`：取 `/menu/routes` → 注册集 = 后端菜单去掉静态已注册 path（防覆盖组件）；`menuRoutes`（新 state）= 后端菜单去掉与 `homepageRouterList` 重复的根——**作为侧栏唯一业务菜单源**。
  - `initRoutes`：`routers = homepageRouterList + menuRoutes`，**`fixedRouterList` 不再进侧栏**——静态演示/业务模块（result、system、user 等）退出侧栏但路由仍注册、URL 可直达（越权访问由 API 拒绝 + 按钮隐藏兜底）。
  - **无需新增加载标志**（评审修正）：guard 沿用 `asyncRoutes.length === 0` 判定，依赖 `transformObjectToRoute` 恒返回 `PAGE_NOT_FOUND` 项使首次构建后非空——这是隐式契约，**改动 `transformObjectToRoute` 的 404 项会导致 guard 重复拉菜单**。
  - `permission-fe.ts` 的按角色过滤方案不采用（记录避免误用）。
- 新增 `v-permission` 指令（`web/src/directives/permission.ts`，main.ts 注册）：值为 perms 码字符串或数组（数组=任一命中即保留）；不满足则从 DOM 移除；超管角色直接放行。
- 接入范围：`web/src/pages/system/{user,role,menu,dept,post,dict,config}` 七个页面的主要操作按钮（新增/编辑/删除/分配），引用的 perms 码与菜单种子一致（共 29 处）。

### S2.4 种子迁移：菜单补齐 + zc1001 可验收

- `script/migrations/0004_rbac_enforce_seed.sql`（幂等，全表 `public.` 限定）：
  - **新建字典管理、参数管理菜单行**（0002/0003 均无其完整菜单行，侧栏切换后否则超管也丢这两项）：C 类型，parent=系统管理，带 path/component/perms（`system:dict:list`、`system:config:list`）；同时修正 id9/id12 遗留 component 与 0002 perms 回填 bug（id12/10 强制为 `system:user:list`/`system:menu:list`，根目录行 perms 清空）；
  - **补齐按钮行**：`system:{user,menu,dict,config}:{add,edit,remove}`（F 类型挂对应菜单，共 12 行）；
  - `system_user_role`：绑定 zc1001 → **既有** `system_default`（dump 已含该角色；缺失时 `NOT EXISTS` 幂等跳过，不静默新建）；
  - `system_role_menu`：为普通角色勾选**可观察子集**——首页 + 用户 + 角色 + 岗位（6 菜单），**仅授予列表 perms、不授予任何按钮行**：既让 zc1001 在这些页面上验收 `v-permission` 按钮隐藏，也避免经 `system:user:add`/`system:role:assign` 自提权（设计修正：原「含其按钮」作废）。

### S2.5 测试与验收边界

- 单测/Handler：菜单过滤三态（超管全量 / 普通角色子集 / 无角色空）+ 未登录；perms 条件五态（超管豁免 admin|system / 命中 / 拒绝 / 通配符 / 匿名）。
- 命令：`./script/check.sh`、`WEB=1 ./script/check.sh` 全绿；eslint/stylelint 改动文件绿。
- curl 验收：未登录 401；超管（skyx）全量菜单与管理接口 200；zc1001 只见子集、`user/list` 200、`menu/list` 信封拒绝；login-only 端点 zc1001 可达。
- 浏览器验收：双角色侧栏差异、zc1001 用户页无新增/行操作按钮、skyx 对应按钮可见、直达页打开且接口拒绝、console 0 error。
- **已知边界（评审记录，均非本次引入或按既有口径沿袭）**：①仅授子未授父时 `buildRouteTree` 从根递归会静默丢弃孤儿子菜单（种子含父目录，当前无影响）；②`FindMenuIDsByUser` 与 `FindRoleCodes/FindPermissionCodes` 同口径不过滤 `user_role.is_remove`（链路自洽）；③`v-permission` 仅 `mounted` 判定，权限在会话内不变，变更需重挂载；④越权返回 HTTP 200 + 信封错误码（项目 response 封装既有行为）。

## [S3] Out of Scope

- 字典关联（状态/类型字典化）——下一 feature `dict-integration`。
- 参数消费（分页默认值、系统名称）——下一 feature `config-consumption`。
- 数据权限（data_scope）查询过滤的行为变更（`system-rbac` 已交付）。
- 登录/OAuth2 会话语义、菜单管理 UI 功能变更。
- 按钮以外的细粒度前端控制（行内操作以外的新 DSL）。
- 既有 `is_remove` 口径统一、孤儿菜单显式报错（见 S2.5 已知边界）。

## Tasks

- [x] T1: `/menu/routes` 按角色过滤（admin 全量/子集/空）+ 路由组挂 `permission.Default()` — acceptance: 未登录 curl 401；admin 返回全量 routes；普通角色返回子集；无角色返回空 (covers: S2.1)
- [x] T2: `/system/*` 逐路由挂 perms 校验 + admin 豁免 + 字典/参数菜单 perms 码补齐 — acceptance: 有 perms 角色调用 200、无 perms 调用 403、admin 200；单测覆盖三态 (covers: S2.2)
- [x] T3: 前端保存 permissions + `v-permission` 指令 + 七个 system 页面按钮接入 — acceptance: `WEB=1 ./script/check.sh` 绿；无权限时对应按钮不渲染；路由渲染仍走后端全量信任路径 (covers: S2.3)
- [x] T4: `0004_rbac_enforce_seed.sql`：zc1001 绑普通角色 + role_menu 子集（幂等） — acceptance: 迁移重复执行成功；zc1001 登录侧栏为约定子集 (covers: S2.4)
- [x] T5: 端到端验收 + 独立评审 — acceptance: check 双连绿；S2.5 验收步骤逐条通过；评审无 CRITICAL (covers: S2.5; depends: T1 T2 T3 T4)
