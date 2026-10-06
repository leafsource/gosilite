---
feature: pro-table-system-config
status: delivered
updated: 2026-06-27
branch: develop
commits: a04f9d3a..working-tree
---

# ProTable 通用查询表格 + 字典/参数管理

## Report

**What was built** — 前端命名四规则写入 coding.md；ProTable（搜索/工具栏/列显隐/分页/刷新/selection-change）；字典+参数迁移 0003、gorm gen、仓储层、Service 唯一性校验、hz IDL 与 Handler、Module 挂载；dict/config 管理页与路由；dept/post 列表迁 ProTable。

**Verification** — `WEB=1 ./script/check.sh` PASS；dict/config 单测通过；curl 字典类型/数据/参数 200、未登录 401、重复 dictType 返回 701000020；评审无 CRITICAL，3 个 MAJOR（唯一键单测、仓储层、dept 搜索）已修。

**Journey log** — 评审指出 Service 内嵌 SQL 违反分层，已抽 dict/config repository；dept 表格关键字需过滤 displayList 而非仅触发事件；ProTable 服务端分页数据 ≤pageSize 时无需 disable-data-page。

## [S1] Problem

中后台列表页各自手写搜索区、工具栏、分页与列控制，system 五页重复且体验不一致；同时缺少字典（固定枚举数据）与参数（动态配置）两块系统维护能力。需要可复用的查询表格组件（搜索、功能区按钮、列显示/排序等），并落地字典/参数完整管理（API + UI）。

## [S2] Design

### S2.1 前端命名规范（强制写入 coding.md）

1. 单文件组件/类：`PascalCase.vue`（如 `ProTable.vue`）
2. 目录主文件：`index.ts` / `index.vue`（导入用目录路径）
3. 接口定义：`camelCase`（如 `dict.ts`、`dictModel.ts`）
4. 资源/样式：`kebab-case`（如 `pro-table.css`）

### S2.2 ProTable 封装

- 路径：`web/src/components/pro-table/index.vue`（目录 + index）+ 类型 `proTable.ts` / 样式 `pro-table.css`（kebab）。
- 能力（中后台标准）：
  1. **搜索区**：`searchFields` 配置或默认插槽；查询/重置；关键字快捷搜索
  2. **功能区**：`toolbar` 插槽（新增、刷新、密度等业务按钮）
  3. **表格**：`columns` + `data` + 分页 + loading
  4. **列控制**：显示/隐藏（checkbox）；列顺序可配置（初始化 order，列表内可调序或按配置排序）
  5. 事件：`page-change`、`refresh`、`selection-change`（预留）
- 基于 TDesign `t-table` / `t-form`；**组件 API 必须先经 TDesign MCP `get-component-docs` 核对**（见 AGENTS.md）。
- 先落地组件与 Story/单页 demo；**system 五页迁移**为可选第二波（若时间盒内完成则迁 dept/post/role 至少 2 页作示范）。

### S2.3 字典管理（后端）

- 表 `system_dict_type`（id, dict_name, dict_type 唯一, status, order_num, remark, del_flag, timestamps）
- 表 `system_dict_data`（id, dict_type 关联字符串, dict_label, dict_value, css_class, list_order, status, is_default, del_flag, timestamps）
- 迁移：`script/migrations/0003_system_dict_param.sql` 幂等 + 种子（如 sys_user_sex、sys_status）
- Service：具体类型 + NewService；错误码 `types/errno` 增补
- HTTP（camelCase，信封 `{code,msg,data}`）：
  - `GET /system/dict/type/list` 分页 keyword
  - `POST/PUT/DELETE /system/dict/type`
  - `GET /system/dict/data/list?dictType=`
  - `POST/PUT/DELETE /system/dict/data`
  - `GET /system/dict/data/options?dictType=` （下拉选项）
- IDL：`idl/framework/system_dict.thrift` → hz 生成；Handler 手写；挂 framework Module，路由 `permission.Default()`

### S2.4 参数管理（后端）

- 表 `system_config`（id, config_name, config_key 唯一, config_value, config_type 系统/业务, remark, status, del_flag, timestamps）
- HTTP：
  - `GET /system/config/list` 分页 keyword
  - `POST/PUT/DELETE /system/config`
  - `GET /system/config/key?key=` 按键取值（运行时）
- 同上：迁移、Service、IDL、Handler、Module 挂载

### S2.5 前端字典/参数 UI（完整管理台）

- 命名：`web/src/api/system/dict.ts` + `dictModel.ts`；`config.ts` + `configModel.ts`
- 页面：`web/src/pages/system/dict/index.vue`（类型 tab/主从或两段表 + 数据项弹层）、`config/index.vue`
- 路由：扩展 `web/src/router/modules/system.ts`（字典管理、参数管理）
- 列表页优先使用 **ProTable**

### S2.6 测试边界

- Service：dict_type 唯一、dict_data 按 type 过滤、config key 唯一
- 前端：`WEB=1 ./script/check.sh`
- 冒烟：curl 字典/参数 CRUD 信封；未登录 401

## [S3] Out of Scope

- system 五页全量 ProTable 迁移（仅示范 2 页）
- 字典/参数缓存层（Redis）与热更新广播
- 导出 Excel、批量导入
- 其他中后台重型能力（审批、打印）

## Tasks

- [x] T1: 前端命名规范写入 `docs/patterns/coding.md`（四条规则）— acceptance: coding.md 含完整四条且与 AGENTS 无冲突 (covers: S2.1)
- [x] T2: ProTable 组件（搜索/工具栏/列显隐/分页/刷新）+ 类型与样式按命名规范 — acceptance: `npm run build` 通过；demo 或 1 页可渲染查询+表格 (covers: S2.2)
- [x] T3: 迁移 0003 字典/参数表 + gorm gen 配置并生成模型 — acceptance: 幂等 SQL 成功；`*.gen.go` 生成 (covers: S2.3 S2.4)
- [x] T4: 字典/参数 Service + errno + 单测（唯一键）— acceptance: `go test` 绿 (covers: S2.3 S2.4; depends: T3)
- [x] T5: IDL + hz + Handler + Module 挂载 — acceptance: curl CRUD 200、未登录 401 (covers: S2.3 S2.4; depends: T4)
- [x] T6: 前端 dict/config API + 两页 UI + 路由（列表用 ProTable）— acceptance: `WEB=1 ./script/check.sh` 通过 (covers: S2.5; depends: T2 T5)
- [x] T7: system 至少 2 页迁 ProTable 示范 — acceptance: dept 与 post 列表使用 ProTable 且 build 过 (covers: S2.2 S2.5; depends: T2)
- [x] T8: 端到端验收 + 独立评审 — acceptance: check 四连绿；评审无 CRITICAL (covers: S2.6; depends: T6 T7)
