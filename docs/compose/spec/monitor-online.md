# Feature Spec: 系统监控 · 在线用户

- status: delivered
- created: 2026-02-07

## Problem

运维缺少「当前谁在线、从哪登录、可立即踢下线」的监控入口。新增顶级菜单**系统监控**（系统管理下方），首个功能**在线用户**：当前系统中活跃用户状态监控。

列表字段（用户拍板清单）：登录名称 / 部门名称 / 客户端 IP / 浏览器 / 操作系统 / 登录时间；行操作仅**强退**。

## 澄清拍板（Gate 已过，2026-02-07）

| 决策点 | 拍板 |
|--------|------|
| 强退粒度 | **会话级**：一行=一次登录（同用户多端多行），只踢选中设备（access denylist + 该会话 refresh 吊销 + 删登记），不影响其他设备 |
| 存量会话 | **不可见可接受**：登记式实现，功能上线前已登录的会话需重新登录才出现 |
| 分页/排序 | 登录时间倒序、每页 10 条（项目惯例）；不加「最后活跃时间」列（字段清单外不扩） |

## Design

### 数据源：Redis 会话登记（唯一可行路径）

当前 Redis 只有 refresh token / 用户 generation / access denylist，无可枚举的活跃 token。新增**单 Hash** `gosilite:auth:online_sessions`：

- field = `sha256(accessRaw)`；value = JSON `{userId, account, deptId, ip, browser, os, loginAt, expireUnix, refreshHash}`。
- `refreshHash = sha256(refreshRaw)`——强退时无需明文即可删 `gosilite:auth:refresh:{refreshHash}`。
- Hash 无 field 级 TTL：**惰性清理**——列表读取时过滤 `expireUnix < now` 的 field 并 HDel（在线量小，HGETALL 可接受）。
- 为此给 `cache.HashOperations` 扩 `HDel`（唯一实现 redis + pipeline 同步补）。

### 链路维护（登录/登出/刷新）

| 时机 | 动作 |
|------|------|
| 登录成功（`LoginResponse`，有 ip/ua） | 查用户 account/deptId + `useragent.Parse` → `PutOnline`；失败仅 hlog 告警，不阻塞登录 |
| 刷新（`Refresh` handler，rotation 换新 access/refresh） | `MoveOnline(oldAccess, newAccess, newRefresh, …)`：删旧 field + 写新 field |
| 登出（`Logout` handler） | `RemoveOnline(accessRaw)`（与既有 Deny/Revoke 并列） |
| 强退 | 读 field → deny `gosilite:auth:access_deny:{hash}` + delete `gosilite:auth:refresh:{refreshHash}` → HDel；找不到 → errno |

- `useragent.Parse`：operlog 中间件的 `parseUA` 上提至 `application/base/useragent`（api→application 合法方向），operlog 改引用。

### 契约（IDL `idl/framework/system_monitor.thrift`，namespace `framework.monitor` → hz）

```
GET  /system/online/list   page,pageSize,userName(模糊) → {list,total}
POST /system/online/force  {id: accessHash} → i32
```

- 行 id = accessHash（十六进制串，天然区分同用户多端）。
- 权限：`rootMw=permission.Default()`；list 挂 `system:online:list`，force 挂 `system:online:force`。
- 内存分页（登录时间倒序）；部门名批量 `dept.List()` 建 map 现查（部门改名实时反映）。

### 菜单（迁移 0015_monitor_online.sql）

- 47 `M` 系统监控（parent=0，sort=3 排系统管理下方，icon `precise-monitor`）
- 4701 `C` 在线用户（parent=47，path `online`，component `monitor/online/index`，perms `system:online:list`）
- 470101 `F` 强退（perms `system:online:force`）；system/admin 角色绑定

### 前端

- `web/src/api/system/online.ts` + model；`web/src/pages/monitor/online/index.vue`：s-table（登录名称模糊搜索），列对齐拍板清单，操作「强退」（popconfirm，`v-permission="'system:online:force'"`）。
- 顶级路由 `web/src/router/modules/monitor.ts`：`/monitor`（LAYOUT，icon PreciseMonitorIcon，orderNo 排系统管理后）→ children `online`。
- 菜单列「部门名称」空值显示 `—`。

## Out of Scope

- 登录日志/操作日志自动清理；并发登录数限制；踢人后前端实时推送（刷新列表即可）。

## Tasks

- [x] T1 spec + 迁移 0015（菜单/权限）
- [x] T2 会话登记 Store + 登录/登出/刷新链路 + useragent 上提
- [x] T3 IDL + hz + Service/Handler + 单测
- [x] T4 前端在线用户页 + 路由菜单
- [x] T5 WEB=1 check + 浏览器验收

## Report

**验收（2026-02-09）**

- `WEB=1 ./script/check.sh` → check OK；新增单测：`TestPageOnline`（过滤/倒序/分页/部门名）、`TestPutAndListOnline`、`TestListOnlineLazyExpires`（惰性过期清理）、`TestMoveOnline`（rotation 迁移 + **登录时间保留**）、`TestForceOnline`（deny+吊销 refresh+摘登记+不影响他端）、`useragent.TestParse`。
- curl 端到端：登录 → 列表出现（部门名/浏览器/IP/登录时间齐全）；`POST /system/online/force` → 该 token 立即 401、列表消失、其他会话保留；被踢后空 id 校验由 service 层覆盖。
- 浏览器：侧栏「系统监控」位于系统管理下方；在线用户页表头=拍板清单；**强退**弹层（popconfirm「确定」）→ toast「已强退 skyx」→ 行消失。会话期间 access 过期由单飞 refresh 自动恢复并跳回原页——refresh→MoveOnline 迁移链路的活证据。
- 过程发现并修复：`MoveOnline` 原会重置登录时间为刷新时刻——已改为沿用原会话 `LoginAt`（refresh ≠ 重新登录）。
- **强退后 refresh 报错修复（用户报障）**：强退会话（含踢自己）后单飞 refresh 打到已删的 refresh token，`RefreshSession` 返回 700000004——该码未注册文案（回落 `DefaultErrorMsg`）且 Refresh handler 把 `err.Error()`（含 code+stack 全量 dump）直接塞进响应 msg。三处修复：① `types/errno/user.go` 注册 `ErrUserSessionInvalidateCode` 文案；② Refresh handler `errors.As` 取 `Msg()` 纯消息；③ 前端 `refreshAccessToken` catch 统一抛 `t('pages.login.expired')`。复现验收：踢自己 → toast「已强退 skyx / 登录已失效，请重新登录」，无 Service Internal Error、无 stack，正确跳登录页。
- 验收笔记：popconfirm 确认按钮文案为「确定」（非「确认」）；v-permission 指令在静态模板位置与 table slot 中均可正常配合 popconfirm（曾误判为指令干扰，实为验收脚本错词 + 重复点击 toggle）。
