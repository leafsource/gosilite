---
feature: role-authorize
status: delivered
updated: 2026-10-07
branch: develop
commits: role-authorize 交付（AuthorizeDrawer + MenuPermTable + PUT/GET /system/role/authorize）
---

# 角色授权（合并「分配菜单 + 数据部门」为「授权」）

## Report

## [S1] Problem

角色页操作列现在是两个独立入口：「分配菜单」（`t-tree` 全选框树，菜单与按钮混在一层）与「数据部门」（另一棵 `t-tree`）。问题：

1. 入口分裂——一次给角色授权要开两个抽屉、点两次保存，语义上却是同一件事。
2. 层级不直观——目录/菜单/按钮混在同一棵树里勾选，看不出「这个角色能进哪些菜单、有哪些功能权限」。
3. 权限粒度不可见——按钮权限（menuType=F，`perms` 权限码）只能靠节点名称猜，没有「权限标识」的明确呈现，也没有与所属菜单、上级菜单的勾选联动。

用户决策：合并为单一「授权」操作；菜单权限区用 EnhancedTable 树表（行=菜单，权限标识列内挂权限 checkbox），**独立组件实现，不走 s-table**，放 `role/components/`。

**第二轮修订（本轮）**：数据对齐改为**后端负责组装**——新增 authorize 快照/整体保存接口，删除旧 4 个 menu/dept 接口；菜单行显示真实菜单图标；「数据部门」页签改为「数据权限」（先选 `sys_data_scope` radio，仅自定义时展示部门树），并成为数据范围的**唯一编辑入口**（角色编辑表单移除该字段）。

## [S2] Design

### S2.1 入口与抽屉骨架

- 角色列表操作列：删除「分配菜单」「数据部门」两个 `t-link`，替换为单个「授权」（`v-permission="'system:role:assign'"`）。
- 抽屉：`t-drawer`，`size="large"`，`destroy-on-close`，标题 `授权：<角色名>`，标题旁展示角色编码/数据范围摘要。
- 抽屉内为自定义分段页签（pill 样式，非 `t-tabs`，满足「可不使用 TDesign、自由发挥」）：**菜单权限** / **数据权限**，切换只改展示、不丢状态。
- 抽屉底部固定操作条：左侧显示统计（菜单/权限/数据范围），右侧「取消 / 保存」——**一次保存**：`PUT /system/role/authorize` 同时提交 menuIds、deptIds、dataScope。
- **滚动模型（第三轮修正）**：不使用 `t-drawer` 的 footer 插槽，但要把自定义操作条钉在可视区底部——`:deep(.t-drawer__body){overflow:hidden}` 关掉抽屉整体滚动，`.auth-drawer` flex 纵向吃满高度：页签区 `flex:none`、内容区 `flex:1 min-height:0`、底部操作条 `flex:none`；菜单权限页只有表格体滚动（见 S2.2 固定表头），数据权限页在自己的区域 `overflow-y:auto` 内滚。
- 抽屉组件：`web/src/pages/system/role/components/AuthorizeDrawer.vue`（业务组件，仅角色页使用）。
- 角色「新增/编辑」表单**移除「数据范围」**字段：新建默认 `dataScope=1`，编辑沿用行上现值不改；数据范围只在授权抽屉里修改（单一入口）。

### S2.2 菜单权限组件 `MenuPermTable.vue`

位置：`web/src/pages/system/role/components/MenuPermTable.vue`；独立实现，**不引用 `s-table`**。

**数据输入（后端已组装，前端零拼装）**

- 输入：`GET /system/role/authorize?id=` 返回的 `RoleAuthorizeResponse`（契约见 S2.4）：菜单树**已把 F 按钮提升为所属 M/C 行的 `perms` 子数组**，行与权限均带 `checked` 布尔态。
- 行节点结构：`{ id, menuName, menuType, icon, checked, children[], perms: PermItem[] }`，`PermItem = { id, menuName, perms, checked }`。
- 前端只做**状态渲染映射**：`checked=true` 的行进 `selectedRowKeys`、`checked=true` 的权限进 `permChecked`——不再做 id 集合拆分、F 提升、属主反查。
- id 归一化：状态集合比较一律 `String(id)`；**保存载荷用原始 id**（数值），后端 `[]int64` 绑定。

**列**（`t-enhanced-table`，全局已注册）

| 列 | 说明 |
|----|------|
| checkbox 列（width 48） | 行勾选，`type: 'checkbox'`，由表格 `selectedRowKeys` 受控 |
| 菜单名称（tree 节点列，`treeNodeColumnIndex: 1`） | **菜单真实图标**：`<t-icon :name="icon" />`（与侧栏 `MenuContent` 同款渲染），`icon` 为空时按类型回退默认图标（M→folder、C→file）+ 名称 |
| 菜单类型 | `sys_menu_type` 字典 tag（与菜单页同款 theme 规则） |
| 权限标识 | 每个权限项 = `checkbox + 权限名称`（**不显示 `xx:xx:x` 权限码**），列内 **2 列 grid**（`repeat(2,minmax(0,1fr))`），不换行堆叠 |
| （布局） | 表格外包 `flex:1 min-height:0` 容器，`:deep(.t-table){height:100%}` + 表格 `max-height="100%"`：触发 TDesign `isFixedHeader`（sticky 表头 + 表体内部滚动），表头以上（工具条）与底部操作条不动 |

tree 配置：`{ childrenKey: 'children', treeNodeColumnIndex: 1, checkStrictly: false }`——`checkStrictly` 必须**显式 false**（TDesign `useTreeSelect.handleSelect` 仅在此字面量下执行父子级联，已核对本地源码）；展开态用受控 `expandedTreeNodes`（过滤/数据变化后重放全展开）。

**联动状态机**

- 状态：`selectedRowKeys`（菜单行勾选）+ `permChecked: Set<string>`（F 权限 id）。不变量：**权限勾选 ⇒ 所属行必在 `selectedRowKeys`**。
- 勾行 → 表格原生级联子孙行；组件在 `onSelectChange` 后做**全树闭包**（把被勾行的全部子孙行补进 keys），保证搜索过滤视图下也落到全量状态；同时**子树内所有行的权限项全勾**。
- 取消行 → 表格级联移除子孙行；对「由勾选变为不勾、且子树内已无任何勾选行」的行，**清空其子树权限项**。（若子树仍有勾选行——TDesign 取消子行时会连带把父行移出 keys——则保留该行权限，仅显示为未勾/半选。）
- 勾权限项 P → 将其所属行补进 `selectedRowKeys`（不强制祖先入 keys，祖先由表格派生勾选/半选态）；取消权限项 P → 仅清该权限，所属行保持。
- 权限项状态直接读 `permChecked`（单元格自渲染 checkbox），不依赖表格选择列。
- 工具条：即时搜索（过滤菜单名称/权限名称/权限码，保留祖先链）、全选、清空、展开全部、收起全部、已选统计（`已选 N/M 菜单 · N 权限`）。
- 空态：树为空或搜索无结果时给占位行。

### S2.3 数据权限 Tab（原「数据部门」）

- 页签名「数据权限」；第一层是 `sys_data_scope` 字典驱动的 **radio**（全部数据/本部门/本部门及以下/仅本人/自定义部门，`value=Number(o.value)`，与角色列表列同源 `useDict('sys_data_scope')`）。
- **仅当 `dataScope === 5`（自定义部门）** 才渲染下方部门树：`t-tree` checkable `value-mode="all"` + 即时搜索 + 全选/清空；部门树勾选态来自 authorize 快照的 `deptIds`，树数据来自 `fetchDeptTree`（模块级缓存）。
- 非自定义时隐藏树；保存仍提交当前树上勾选的 `deptIds`，由服务端在 `dataScope≠5` 时清空（见 S2.4）。
- 页签计数显示当前 dataScope 的字典 label（自定义时附部门数）。

### S2.4 后端授权契约（新增 authorize，删旧 4 接口）

**契约（IDL `idl/framework/system.thrift`，hz 生成）**

```thrift
struct RolePermItem   { 1: i64 id, 2: string menuName, 3: string perms, 4: bool checked }
struct RoleAuthMenuNode {
  1: i64 id, 2: i64 parentId, 3: string menuName, 4: string menuType,
  5: string icon, 6: i32 orderNum, 7: bool checked,
  8: list<RolePermItem> perms, 9: list<RoleAuthMenuNode> children
}
struct RoleAuthorizeResponse { 1: i32 dataScope, 2: list<RoleAuthMenuNode> menus, 3: list<i64> deptIds }
struct RoleAuthorizeSaveRequest { 1: i64 roleId, 2: list<i64> menuIds, 3: list<i64> deptIds, 4: i32 dataScope }

RoleAuthorize     GET  /system/role/authorize   // IdRequest → RoleAuthorizeResponse
RoleAuthorizeSave PUT  /system/role/authorize   // RoleAuthorizeSaveRequest → i32
```

同时**删除**：`RoleAssignMenu`、`RoleAssignDept`、`RoleMenus`、`RoleDepts` 四个操作及 `RoleMenuRequest`/`RoleDeptRequest` 结构。

**GET 组装（服务端，前端零处理）**

- 全量非删除菜单树；F 按钮提升为最近 M/C 祖先行的 `perms`（无 M/C 祖先的 F 忽略），按 `orderNum` 排序；行带 `icon`。
- `checked`（行/权限）= 对应 id ∈ 角色 role_menu 集合；**归一化**：权限 checked ⇒ 所属行 checked（服务端保证 S2.2 不变量，前端不再属主反查）。
- 附带 `role.dataScope` 与角色 `deptIds`。

**PUT 保存（服务端）**

1. `menuIds` **补全祖先闭包**（按非删除菜单的 parentId 链上溯）后整单替换 role_menu——前端只提交选中行+勾选权限，不再自算祖先；
2. 更新 `role.data_scope = dataScope`；
3. `dataScope == 5` 时整单替换 role_dept 为 `deptIds`，否则清空 role_dept；
4. 权限挂载：GET/PUT 均 `system:role:assign`（middleware.go 对应生成函数）。

**前端对接**

- `web/src/api/system/role.ts`：删除 `assignRoleMenu/assignRoleDept/fetchRoleMenus/fetchRoleDepts` 及 `RoleMenuParams/RoleDeptParams`；新增 `fetchRoleAuthorize(id)` / `saveRoleAuthorize(data)`，类型按上述 IDL（camelCase）。
- `AuthorizeDrawer`：打开只调一次 GET（+ 缓存的部门树）；保存只调一次 PUT；失败报错并保持抽屉打开，成功关闭。

### S2.5 验收边界

- `WEB=1 ./script/check.sh` 全绿（`vue-tsc` + `vite build` + go test/vet/build）。
- 操作列仅剩一个「授权」入口；抽屉含双 Tab、一次保存（单次 PUT）。
- 联动行为与 S2.2 状态机一致；菜单行显示真实图标（空回退）；数据权限仅自定义时出现部门树；编辑表单无「数据范围」。
- 旧 4 接口在 IDL、路由、handler、前端 API 中均已删除，无残留引用。
- 环境可用时登录冒烟（`skyx/Admin@123`）走通授权保存闭环。
- TDesign MCP 本会话未暴露工具：`t-*` API 以 `web/node_modules/tdesign-vue-next` 本地类型与源码核对（`TableTreeConfig`、`useTreeSelect` 级联、`indeterminateSelectedRowKeys` 已核对），与 menu-s-table 交付时同一回退路径。

## [S3] Out of Scope

- 数据范围的**执行语义**（`VisibleDeptIDs` 等数据过滤计算）、字典 `sys_data_scope` 种子本身。
- 角色编辑表单其余字段、角色列表列定义。
- `s-table` / 共享组件扩展；其他业务页面的授权交互。
- F 按钮节点作为表格行展示；表结构迁移（role_menu/role_dept 复用现表）。
- 前端单测基建（`web` 无测试运行器，验收以 build/lint 为准）。

## Tasks

- [x] T1: `MenuPermTable.vue` — 树表 + 权限标识列 + 联动状态机 + 工具条 — acceptance: `cd web && npm run build` 通过，组件位于 `web/src/pages/system/role/components/`，未引用 `s-table` (covers: S2.2)
- [x] T2: `AuthorizeDrawer.vue` 与角色页接入 — 双 Tab 骨架、部门树迁入、单次保存、操作列合并为「授权」 — acceptance: `npm run build` 通过，操作列仅一个授权入口 (covers: S2.1; depends: T1)
- [ ] T4: 后端 authorize 契约 — IDL 增删 → hz update → GET/PUT 组装与保存实现、删旧 handler/service 方法、权限挂载 — acceptance: `./script/check.sh` 全绿；`GET/PUT /system/role/authorize` 可用，旧 4 路由不存在 (covers: S2.4)
- [ ] T5: 前端对接修订 — role.ts 新契约、MenuPermTable 接快照+真实图标、AuthorizeDrawer 数据权限页签与单次保存、编辑表单移除数据范围 — acceptance: `npm run build` 通过，无旧接口引用 (covers: S2.1, S2.2, S2.3, S2.4; depends: T4)
- [ ] T6: UI 修正 — 权限标识去权限码 + 2 列 grid；抽屉固定布局（表头/页签/底栏不动，仅表体滚动） — acceptance: `npm run build` 通过；冒烟中滚动表格时表头与底部「取消/保存」不位移，权限标识列每行两列且无权限码 (covers: S2.1, S2.2; depends: T5)
- [ ] T3: 验收与评审 — acceptance: `WEB=1 ./script/check.sh` 全绿，独立评审无 critical (covers: S2.5; depends: T6)
