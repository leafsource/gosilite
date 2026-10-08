---
feature: notice
status: delivered
updated: 2026-10-08
branch: develop
commits: working-tree
---

# 通知公告：系统通知公告信息发布维护

## [S1] Problem

系统缺少通知公告的发布与维护能力：管理员无法在后台录入、发布、撤回面向全员的
通知/公告；布局铃铛（`layouts/components/Notice.vue`）目前只消费前端 mock 数据，
没有真实数据源。

目标：交付「通知公告」管理功能——列表（关键字 + 类型/状态筛选）、新增、编辑、
删除、批量删除/批量发布/撤回、只读查看，权限与菜单接入既有 RBAC。

## [S2] Design

### S2.1 数据与迁移（`script/migrations/0011_notice.sql`，幂等、`public.` 限定）

新表 `public.system_notice`：

| 列 | 类型 | 说明 |
|----|------|------|
| id | BIGSERIAL PK | |
| title | VARCHAR(128) NOT NULL | 标题 |
| notice_type | SMALLINT NOT NULL DEFAULT 1 | 字典 `sys_notice_type`：1 通知 / 2 公告 |
| content | TEXT NOT NULL DEFAULT '' | 正文（纯文本） |
| status | SMALLINT NOT NULL DEFAULT 0 | 字典 `sys_notice_status`：0 已发布 / 1 已撤回 |
| publisher | VARCHAR(64) NOT NULL DEFAULT '' | 发布人（登录态写入） |
| del_flag | SMALLINT NOT NULL DEFAULT 0 | 0 正常 / 2 删除 |
| created_at / updated_at | TIMESTAMP DEFAULT NOW() | |

- 字典种子（挂「系统」分组，仿 0010 模式）：`sys_notice_type`（通知/公告）、
  `sys_notice_status`（已发布/已撤回）。
- 菜单种子：id **45**（44 已被组织架构占用）C 菜单「通知公告」path=`notice`
  component=`system/notice/index` perms=`system:notice:list`，sort=9；F 按钮
  4501/4502/4503（`system:notice:add/edit/remove`）；`system`/`admin` 角色
  role_menu 绑定（仿 0002）。
- 状态语义：`0 已发布 / 1 已撤回`——新建可直接选状态；撤回/发布走编辑或批量
  操作，不单设工作流。

### S2.2 后端（路由必须 hz 生成——AGENTS 已移除「可跳过 hz」条款，本机 hz v0.9.7 / thriftgo 0.4.5）

- **契约 IDL** `idl/framework/system_notice.thrift`（`namespace go framework.notice`）：
  `NoticeListRequest/NoticeDTO/NoticeListResponse/NoticeSaveRequest/NoticeIdRequest` +
  `service SystemNoticeController` 五方法；`hz update -I idl -idl … -customize_package
  idl/package.yaml` 生成 `api/model/framework/notice/system_notice.go`（请求/响应 DTO）
  与 `api/router/framework/notice/system_notice.go`（`Register`）+ `middleware.go`。
- **Handler** `api/handler/framework/notice/system_notice_controller.go` 只做
  绑定生成类型 → 调 Service → 回包；`convert.go` 做 dal model ↔ 生成 DTO 映射。
- **权限** 生成后仅改 `middleware.go`：`rootMw` = `permission.Default()`，
  `_listMw/_getMw` = `system:notice:list`、`_createMw` = `system:notice:add`、
  `_updateMw` = `system:notice:edit`、`_removeMw` = `system:notice:remove`：

| 方法 | 路径 | 权限码 |
|------|------|--------|
| GET | `/system/notice/list` | `system:notice:list`（page/pageSize/keyword/noticeType/status） |
| GET | `/system/notice` | `system:notice:list`（id） |
| POST | `/system/notice` | `system:notice:add` |
| PUT | `/system/notice` | `system:notice:edit` |
| DELETE | `/system/notice` | `system:notice:remove`（id） |

- **分层**：`handler → service/notice → dal/repository/notice_repository.go`（接口 +
  GORM，先例 config 仓储）；模型 `dal/model/system_notice.gen.go` 经
  `go run ./script/codegen` 生成（codegen 配置登记 `system_notice`，
  `repository:false, service:false` 跳过模板骨架）。
- **Service**：`Create` 校验标题非空、publisher 取 `ctxutil.GetLoginUser`；`Update`
  先查存在（`ErrSystemNoticeNotFound`）、不改 publisher；`Delete` 软删；`Page`
  keyword LIKE title/content，noticeType/status 可选过滤（nil=全部）。
- **错误码**：`types/errno/system.go` 新增 `ErrSystemNoticeNotFound = 701000028`。
- **响应**：`R.SuccessWithData` 信封 `{code,msg,data}`；时间列输出
  `2006-01-02 15:04:05` 字符串。

### S2.3 前端

- `web/src/api/system/notice.ts` + `noticeModel.ts`（camelCase，envelope 对齐）。
- `web/src/pages/system/notice/index.vue`：
  - `s-table`：关键字（标题/正文）、`searchFields` 类型/状态 select（字典 options，
    查询后生效）；列：标题 / 类型(`s-tag-dict`) / 状态(`s-tag-dict`) / 发布人 /
    创建时间 / 操作；
  - 工具栏「新增公告」`v-permission="system:notice:add"`；`#batch`：批量删除、
    批量发布、批量撤回；
  - 表单 **4 个 `t-form-item` ≤5 → `s-form-dialog`**（标题 / 类型 `s-select-dict` /
    状态 `s-radio-dict` / 正文 `t-textarea`）；
  - 「查看」只读 `t-dialog`（写码前 TDesign MCP `get-component-docs` 核对）。
- 静态路由 `router/modules/system.ts` 增 `notice` → `SystemNotice`（侧栏由后端菜单
  过滤下发，component=`system/notice/index` 动态解析到页面）。
- 字典未加载时 `useDict` 硬编码兜底（dict.ts `DEFAULT_FALLBACK` 补两类型）。

### S2.4 测试与验收

- `service_test.go`：空标题创建报 `ErrSystemInvalidParam`、publisher 写入登录态、
  更新不存在 id 报 `ErrSystemNoticeNotFound`、删除委托仓储。
- 迁移 0011 双跑幂等；`./script/check.sh` + `WEB=1 ./script/check.sh` 全绿。
- 冒烟：登录后 `GET /system/notice/list` 200；无权限角色 403；按钮 `v-permission`
  隐藏。

## [S3] Out of Scope

- 富文本/Markdown 渲染、附件、@人员/定向推送。
- 已读回执、未读数、与布局铃铛组件的真实数据对接（后续可复用本表）。
- 定时发布/到期自动撤回、审批流。

## Tasks

- [x] T1: 规划与 spec — acceptance: 本文件契约/表/路由/权限码定稿 (covers: S1 S2)
- [x] T2: 后端 0011 迁移 + codegen 模型 + repo/service/handler/router/module + errno + 单测 — acceptance: 双跑幂等；go test 绿；登录态 CRUD 冒烟 (covers: S2.1 S2.2 S2.4)
- [x] T3: 前端 api + 页面 + 静态路由 + 字典兜底 — acceptance: 页面可增删改查/批量/查看；按钮权限生效 (covers: S2.3)
- [x] T4: 验收 — acceptance: `WEB=1 ./script/check.sh` 全绿；迁移双跑 0 变更 (covers: S2.4)
- [x] T5: AGENTS/SKILL 移除「跳过 hz」条款，改为路由必须 hz 生成 — acceptance: 全仓无该条款残留；规范写入禁止手写 Register (covers: S2.2)
- [x] T6: notice 迁移 hz（IDL + hz update + 生成模型/路由 + middleware 权限） — acceptance: 手写 register.go 删除；curl 冒烟与 401 通过 (covers: S2.2)
- [x] T7: leader 路由迁移 hz（system.thrift 补 DeptLeader） — acceptance: 删 leader.go/dept_leader.go；PUT /system/dept/leader 200 且 leaderId 落库 (covers: S2.2)
- [x] T8: global 免登录路由迁移 hz（独立 IDL/独立 Register/rootMw=nil） — acceptance: 匿名 global 200、匿名 key 401、登录 key 200 (covers: S2.2)
- [x] T9: 全量验收 — acceptance: `WEB=1 ./script/check.sh` 全绿；api/router 下无手写 Register (covers: S2.4)

## Report

**What was built** — 迁移 `0011_notice.sql`（`system_notice` 表 + `sys_notice_type`/`sys_notice_status` 字典 + 菜单 45/F 按钮 4501-4503 + system/admin role_menu，幂等双跑）；gorm gen 登记 `system_notice`（`repository:false, service:false`，只出 model/query）；分层 `handler(notice) → service/notice → dal/repository/notice_repository`，仓储 `Save` 为选择性更新（不触碰 publisher/created_at，规避整行 Save 清零）；错误码 `ErrSystemNoticeNotFound=701000028`；前端 `api/system/notice(.ts|Model.ts)` + `pages/system/notice/index.vue`（s-table 关键字+类型/状态筛选、`#batch` 批量删除/发布/撤回、4 字段 s-form-dialog、t-dialog 只读查看）+ 静态路由 `SystemNotice` + `dict.ts` 两类型兜底。**返工（hz 强制生成）**：按用户拍板移除 AGENTS/SKILL「纯手写系统管理路由可跳过 hz」条款后，notice 走 `idl/framework/system_notice.thrift` → `hz update`（hz v0.9.7/thriftgo 0.4.5）生成 model/router，handler 改用生成 DTO，权限只落 `middleware.go`；同批把存量手写路由全部收敛——leader 并入 `system.thrift`（`DeptLeader` 方法）、global 拆独立 `idl/framework/system_config_global.thrift`（独立 Register，`rootMw=nil` 保持匿名），删除 `register.go`/`leader.go`/`global.go`/`dept_leader.go`/`config_global.go` 五个手写文件。

**Verification** — 迁移双跑 exit 0 第二遍 0 行；`go test ./application/framework/service/notice/` 7 用例 PASS（空标题/publisher 登录态/NotFound/删除委托）；curl 冒烟（手写路由版）：登录→create/blank-title(701000001)/list 过滤(type+status+keyword)/get missing(701000028)/update 回包完整行/delete；**hz 迁移后复验**：notice 五端点 200 + 未登录 401、`PUT /system/dept/leader` 200 且 `leaderId=51` 落库进 tree、匿名 `GET /system/config/global` 200 含 `sys.app.name`、匿名 `/key` 401、登录 `/key` 200、匿名 dict options 401；`WEB=1 ./script/check.sh` **check OK**；改动文件 eslint exit 0；浏览器端到端（skyx，IAB）：登录 → 侧栏「系统管理」弹出菜单含「通知公告」（notification 图标）→ 新增「系统维护通知」→ 保存成功回列表倒序置顶 → 查看弹窗正文换行保留 → 编辑切「已撤回」保存（创建时间/发布人保留）→ 关键字「维护」筛选、状态「已发布」筛选、重置 → 删除测试行 → console 无新增 error（仅存量 vue-router deprecation / crypto-trade 菜单警告）。

**Journey log** — ①菜单 id 44 已被组织架构占用，通知公告用 45（0011 里写明）；②eslint `consistent-type-definitions` 自动把 `type` 修成 `interface` 后，`Record<string,unknown>[] as NoticeDTO[]` 触发 TS2352（interface 无隐式索引签名，type 有）——批量槽位改收 `unknown[]` 内部一次断言，模板免 cast；③TDesign radio 的原生 input `setChecked` 不可点，点可见标签即可；④Update 回包若直接回显请求体会缺 publisher/时间列，改为落库后重查回完整行；⑤`notification` 图标名经 `search-icon` 核实存在，菜单种子一次通过；⑥会话中出现的「gosilite框架介绍」行非本流程所建，保留未动；⑦**hz 本机实为 `~/go/bin` 已装**（此前 PATH 未含导致误判「无 hz」）——结论：不存在任何跳过 hz 的正当理由，条款已删；⑧hz update 对存量文件是**追加式**（middleware.go 只补新 mw 函数、权限全保留；handler 只在目标 controller 文件缺函数时追加骨架），但骨架不带业务逻辑且引用未导入的 `consts`——必须立即用真实实现替换；⑨global 的匿名约束用「独立 IDL 文件 → 独立 Register → rootMw=nil」解决，不再依赖手写路由。
