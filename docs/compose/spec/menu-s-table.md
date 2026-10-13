---
feature: menu-s-table
status: delivered
updated: 2026-10-07
branch: develop
commits: 95d5a009 # base；实现与本文档均在工作树交付（见 Report），收尾提交待用户决定
---

# 菜单管理接入 s-table（树形模式）

## Report

**What was built** — `s-table` 新增 `tree`（树形配置）与 `paginated`（分页开关，默认 true）两个可选 prop：传入 `tree` 时内层表格经 `<component :is>` 切换为 `EnhancedTable`，否则保持 `t-table`，现有 5 个调用方零改动。菜单管理页列表区整体迁入 `s-table`：树形配置、无分页、无选择框，关键字筛选从自建输入框迁到 s-table 搜索区（查询后生效），「新增菜单」按钮置于 `#toolbar`，原 `menuType/visible/status/op` 单元格插槽原样透传；页面不再有自建工具条与裸 `t-enhanced-table`。菜单新增/编辑表单从 `s-form-dialog`（10 字段对话框）迁至 `s-form-drawer`，字段与校验不变。评审后补充：树节点列（`treeNodeColumnIndex` 默认第 0 列）在列设置面板中不可隐藏、且始终强制可见，防止用户自损树展开入口。

**Verification** — `WEB=1 ./script/check.sh` → `check OK`（go test / go vet / go build / web npm run build 全绿，跑过两轮：首轮交付 + 列保护修复后复跑）；`git diff --name-only` 结合会话基线确认 s-table 调用方页面文件未被本次改动（config/dict/post/role 的在途修改、organization 的 untracked 均先于本会话存在）。独立评审结论：spec 合规 T1–T4 全过，无 critical/major；2 个 minor（调用方零改动的直接证据、列设置隐藏树节点列）均已处理，2 个 nit（本文档状态、`dialogVisible` 命名沿用在途改动）。

**Journey log** —
1. 结构性障碍只有两点（树形 + 无分页），因此选择扩展 s-table 而非样式对齐或左树右表重做——一次性让 6 个列表页 shell 完全一致。
2. 内层表格用 `<component :is>` 条件切换而非恒用 EnhancedTable：保 5 个现调用方零行为变化，双路径共用同一模板不重复插槽透传。
3. 工作区偏离默认：develop 主线 + 大量在途未提交改动是本次依赖（s-form-dialog untracked、menu 页已改），经用户批准直接在当前检出开发，未开 worktree。
4. TDesign MCP 本会话未暴露工具，`t-*` API 改以 `web/node_modules/tdesign-vue-next` 本地类型定义核对（`Table`/`EnhancedTable` 导出、`tree` prop、drawer confirm 不自动关闭均已验证）。
5. `defaultExpandAll` 在数据刷新后由 `tExpandedTreeNode` 重放，搜索过滤后树保持全展开——与原直连 EnhancedTable 行为一致。

## [S1] Problem

系统管理 6 个列表页中，config / dict / organization / post / role 均已使用 `s-table` 作为列表 shell（搜索区、工具栏、列设置、分页统一），唯独菜单管理页仍为自建工具条 + 裸 `t-enhanced-table`，与其他系统页 UI 不一致。菜单页是树形表格（`t-table` 无 tree 能力）且为全量树数据、无服务端分页，这两点是它未接入 `s-table` 的结构性原因。此外菜单新增/编辑表单用 `s-form-dialog` 承载 10 个字段，违反 frontend-dev skill「>5 字段必须用抽屉」的规则。

用户已确认决策：扩展 `s-table` 支持树形而非只统一样式；顺带将表单迁到 `s-form-drawer`；在当前 develop 检出直接开发（不开 worktree，因在途未提交改动是本次工作的依赖）。

## [S2] Design

### S2.1 s-table 扩展（共享组件）

- 新增 prop `tree?: STableTreeConfig`，其中 `STableTreeConfig = { childrenKey?: string; treeNodeColumnIndex?: number; defaultExpandAll?: boolean }`（结构对齐 `TdEnhancedTableProps['tree']`）。
- 新增 prop `paginated?: boolean`，默认 `true`；为 `false` 时不向内层表格传 `pagination`（用于全量树/无服务端分页页面）。
- 内层表格通过 `<component :is>` 切换：传入 `tree` 时渲染 `EnhancedTable`（从 `tdesign-vue-next` 导出），否则渲染原有 `t-table`。模板其余部分（搜索区、工具栏、插槽透传）不拆双份，现有 5 个调用方零改动、行为不变。
- 类型定义同步更新到 `sTable.ts`。

### S2.2 菜单页列表区迁移

- 用 `s-table` 替换页面自建 toolbar + `t-enhanced-table`：
  - `:tree="{ childrenKey: 'children', treeNodeColumnIndex: 0, defaultExpandAll: true }"`；
  - `:paginated="false"`、`:selectable="false"`（菜单无批量操作需求）；
  - `keyword-placeholder="菜单名称/权限标识"`。
- 客户端筛选迁入 s-table 搜索区，采用「查询后生效」模式（skill 模式 A）：`@search` 取 `{ keyword }` 写入 `filterText`，`@reset` 清空；现有 `filteredRows` 树形过滤逻辑复用不变（原「输入即筛」改为点查询生效）。
- `#toolbar` 插槽放「新增菜单」按钮，保留 `v-permission="'system:menu:add'"`。
- 列定义与 `menuType` / `visible` / `status` / `op` 单元格插槽原样透传（s-table 已全插槽转发至内层表格）。
- 删除确认 `t-dialog` 保持不变；页面卡片容器样式保留。

### S2.3 表单迁 s-form-drawer

- `s-form-dialog` → `s-form-drawer`（`size` 用默认 `large`），字段、`v-model:visible`、`:loading`、`@confirm` 语义与现有 `s-form-drawer` 用法（dict / organization 页）一致。
- 表单内容与校验不变：`menuName` 必填校验、保存成功后刷新树并关抽屉。

### 验收边界

- `WEB=1 ./script/check.sh` 全绿。
- 现有 5 个 s-table 调用方文件不被修改。

## [S3] Out of Scope

- 后端接口、菜单树数据结构、权限码。
- 菜单批量选择/批量操作（明确不开 selectable）。
- organization 式「左树右表」交互重做。
- 其他页面表单 dialog/drawer 复查；`s-form-dialog` 组件本身的去留。
- s-table 服务端分页与树形组合场景（当前无需求）。

## Tasks

- [x] T1: s-table 新增 `tree` / `paginated` props，内层表格按 tree 切换 EnhancedTable — acceptance: `cd web && npm run build` 通过，且 config/dict/organization/post/role 页面文件零改动（covers: S2.1）
- [x] T2: 菜单页列表区迁移 s-table — acceptance: 页面中无自建 toolbar 条、无直接 `t-enhanced-table`；树默认全展开，搜索区关键字「查询/重置」生效，`#toolbar` 含带权限的「新增菜单」（covers: S2.2; depends: T1）
- [x] T3: 菜单表单迁 s-form-drawer — acceptance: 新增/编辑在抽屉打开，10 个字段完整，必填校验与保存刷新行为不变（covers: S2.3; depends: T2）
- [x] T4: 验收 — acceptance: `WEB=1 ./script/check.sh` 输出 `check OK`（covers: S2; depends: T2, T3）
