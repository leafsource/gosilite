---
feature: system-rbac
status: delivered
updated: 2026-06-27
branch: develop
commits: a04f9d3a..working-tree
---

# 系统 RBAC：用户 / 部门 / 岗位 / 菜单 / 角色

## Report

**What was built** — 五域系统管理台：部门树/岗位/用户/菜单/角色（数据范围 + 角色菜单分配）。迁移 `0002_system_rbac.sql`；gorm gen 生成模型；`idl/framework/system.thrift` + hz 生成路由；Handler/Service 手写业务；framework Module 挂载；权限源统一为 `system_role_menu` → `system_menu.perms`；前端五页完整 UI。

**Verification** — `./script/check.sh` PASS；`WEB=1 ./script/check.sh` PASS；curl 登录 GetInfo 返回 roles/permissions；`/system/*` 未登录 401、已登录 200；独立评审 4 CRITICAL 修复后复审全部 FIXED、无新 CRITICAL。

**Journey log** — 评审抓到权限双写分裂（role_permission vs role_menu），已统一 role_menu；数据权限改为 fail-closed；部门环检测改用 DB ancestors；路由根兼容 parent 0 与 -1。

## [S1] Problem

基础框架已具备登录会话与 Module 插件，但管理台缺少完整 RBAC 能力：无法配置组织架构（部门树 + 数据权限）、岗位、菜单与按钮权限标识、角色按机构的数据范围划分；用户仅能登录，不能做系统运营配置。现有 `user/role/menu` 表与服务仅覆盖骨架，需按主流管理台（RuoYi 类）标准重建，**不被旧实现绑死**。字典与参数管理不在本 feature（见 `system-config` 后续 spec）。

## [S2] Design

### S2.1 领域与数据模型

沿用 gosilite：`application/framework` 单 Module，**具体 `Service` + `NewService`**，Repository 接口 + GORM gen，错误码 `types/errno`，信封 `{code,msg,data}`。

| 表 | 职责 | 关键字段 |
|----|------|----------|
| `system_dept` | 部门树 | `id, parent_id, ancestors, dept_name, order_num, status, del_flag` |
| `system_post` | 岗位 | `id, post_code, post_name, order_num, status` |
| `system_user` | 用户（增强） | 保留现有字段，**增加** `dept_id, post_ids`（或关联表 `system_user_post`） |
| `system_menu` | 菜单/按钮 | `parent_id, menu_name, menu_type(M目录/C菜单/F按钮), path, component, perms, icon, order_num, visible, status` |
| `system_role` | 角色（增强） | `role_name, role_key, order_num, status, data_scope(1全部/2本部门/3本部门及以下/4仅本人/5自定义)` |
| `system_user_role` | 用户角色 | 已有，保留 |
| `system_role_menu` | 角色菜单 | 替代语义模糊的 `system_role_permission`（迁移期可双写/改名） |
| `system_role_dept` | 自定义数据范围 | `role_id, dept_id`（`data_scope=5` 时生效） |

- 部门 `ancestors` 存逗号分隔祖先 id（如 `0,100,101`），便于「本部门及以下」查询。
- 逻辑删除：`del_flag`（0 正常 / 2 删除）与主流对齐；status：0 启用 / 1 停用。
- ID：沿用现有雪花/数据库 id 策略与 `system_*` 既有约定，不另起一套。

**迁移**：`script/migrations/0002_system_rbac.sql`（幂等 DDL + 菜单种子数据）。不改 Navicat 整库 dump 的历史段落，新表以 migrations 为准。

### S2.2 权限语义

- **功能权限**：菜单 `perms` 字符串（如 `system:user:list` / `system:user:add`），按钮类型 `F` 挂在菜单下；角色勾选菜单树 → `system_role_menu`。
- **数据权限**：查询列表接口按当前用户角色 `data_scope` 过滤部门可见集；`data_scope=5` 用 `system_role_dept`；管理员（超管 `role_key=admin`）不滤。
- 鉴权中间件沿用 `permission.Default()` + `Authorizator` 装载 `LoginUser.Permissions/Roles`；`GetInfo`/`GetProfile` 返回角色与权限码供前端。
- Handler 只绑定/调 Service/回包；部门树重建（改 parent）在 Service 事务内更新子树 `ancestors`。

### S2.3 API 契约（HTTP，camelCase JSON）

统一前缀无 `/api`（由 Vite 代理剥前缀）。全部管理接口：登录 + 按 `perms` 校验（本阶段至少校验已认证；perms 拦在路由中间件按 `permission.Permissions(...)` 挂载）。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/system/dept/tree` | 部门树 |
| POST/PUT/DELETE | `/system/dept`… | 增删改（删=逻辑删，禁止删有子部门/用户） |
| GET | `/system/post/list` | 分页 |
| POST/PUT/DELETE | `/system/post`… | |
| GET | `/system/user/list` | 分页 + 姓名/部门/状态筛选 + 数据权限 |
| POST/PUT/DELETE | `/system/user`… | 创建含角色/岗位 |
| PUT | `/system/user/status` | 启用/停用 |
| GET | `/system/user/{id}/roles` | 已选角色 |
| GET | `/system/menu/tree` | 菜单树（管理用，全量） |
| GET | `/system/menu/routes` | 当前用户可见路由（替换/兼容现有 `/menu/routes`） |
| POST/PUT/DELETE | `/system/menu`… | |
| GET | `/system/role/list` | 分页 |
| POST/PUT/DELETE | `/system/role`… | |
| PUT | `/system/role/menu` | 分配菜单（ids） |
| PUT | `/system/role/dept` | 分配数据范围部门（ids） |
| GET | `/system/role/{id}/menus` | 已选菜单 ids |

**IDL**：新增 `idl/framework/system/*.thrift`（dept/post/user admin/menu admin/role admin），`hz update` 生成；与现有 `user.thrift` 认证接口并存，旧 `LoginResponse` 类残留不再扩展。

**错误码**：`types/errno` 新增 dept/post/menu/role/config 段；统一「父节点存在子级不可删」「岗位编码重复」「角色标识重复」等。

### S2.4 前端（完整管理台 UI）

- 路由模块：`web/src/pages/system/{dept,post,user,menu,role}/` + `web/src/router/modules/system.ts`。
- 列表：TDesign `t-table` 分页；筛选区；新增/编辑 `t-drawer` 或 `t-dialog` 表单。
- 部门：`t-tree` 选择器 + 侧栏树联动列表。
- 菜单：树表 + 类型（目录/菜单/按钮）；权限标识输入；角色分配用 `t-tree` 勾选（`checkable`）。
- 角色：列表 + 数据范围 `t-radio-group`（全部/本部门及以下/仅本人/自定义）+ 自定义部门弹树。
- 用户：列表 + 分配角色/岗位。
- API 层：`web/src/api/system/*.ts`，envelope 按 `{code,msg,data}`；权限按钮用现有权限 store / `v-permission` 约定。
- 导航：侧栏出现「系统管理」分组（部门/用户/角色/菜单/岗位）。

### S2.5 插件与装配

- 全部落在既有 `framework` Module：`Init` 增加新 Service；`RegisterRoutes` 注册 `api/router/framework/system/**`（或按子包 Register）。
- 不新增第二套装配方式；不破坏 `application.Module` 契约。

### S2.6 测试边界

- Service：部门树改父后 ancestors 正确；删部门约束；数据权限 SQL 过滤（角色 2/3/4/5）；菜单 perms 排重。
- Session 相关既有测试保持通过。
- 验收：`./script/check.sh` 全绿 + `WEB=1` 前端 build；本地起服务后按「种子管理员」能进系统管理页。

## [S3] Out of Scope

- 字典管理、参数管理（下一 feature `system-config`）。
- 登录/OAuth2 会话语义变更。
- 组织架构与人力系统集成、审批流。
- 精细到按钮的前端指令框架之外的新 DSL。
- 移动端管理台。

## Tasks

- [x] T1: 迁移 SQL：`system_dept`/`system_post`/`system_user_dept(post)`/`system_role_menu`/`system_role_dept` + menu/role 增强列 + 种子数据 — acceptance: 在空/现有库执行幂等成功，表与种子可查询 (covers: S2.1)
- [x] T2: gorm gen 模型与 Repository（dept/post/user_admin/menu_admin/role_admin）— acceptance: `go build` 通过，repo 单测覆盖树查询与分页 (covers: S2.1; depends: T1)
- [x] T3: 领域 Service：部门树/约束、岗位、用户 CRUD+角色岗位、菜单树+perms、角色 CRUD+菜单/部门分配、数据范围过滤器 — acceptance: 服务测试绿，含改父/禁止删父/数据范围用例 (covers: S2.1 S2.2; depends: T2)
- [x] T4: 系统管理 Handler + 手写路由挂载（`api/handler/framework/system` + `api/router/framework/system` 注册进 framework Module），权限中间件 — acceptance: `curl` 各 CRUD 信封正确；未登录 401 (covers: S2.3; depends: T3)
- [x] T5: framework Module 注册新 Service 与 RegisterRoutes；`/system/menu/routes` 与用户 perms 下发 — acceptance: 登录后 GetInfo 含 roles/perms；路由可达 (covers: S2.5 S2.2; depends: T4)
- [x] T6: 前端 system API + 路由模块 + 五页完整 UI（树、分配、数据范围）— acceptance: `WEB=1 ./script/check.sh` 通过；浏览器可完成建部门→建用户→分角色→勾菜单闭环 (covers: S2.4; depends: T5)
- [x] T7: 错误码注册、AGENTS/skill 文档补系统管理约定 — acceptance: 无散落魔法码；文档含 system 模块指引 (covers: S2.3 S2.6; depends: T4)
- [x] T8: 端到端验收与独立评审 — acceptance: `./script/check.sh` 与 `WEB=1` 全绿；评审无 CRITICAL (covers: S2.6; depends: T6 T7)
