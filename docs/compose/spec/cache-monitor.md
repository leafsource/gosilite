# Feature Spec: 系统监控 · 缓存监控

- status: delivered
- created: 2026-02-10

## Problem

运维需要查看 Redis 里业务键、值，并按需删除。对照三栏效果图，但 gosilite **无 RuoYi 式逻辑 CacheManager**，不能写死假名字。

## 澄清拍板

| 决策点 | 拍板 |
|--------|------|
| 命令统计 | **不做** |
| 权限 | **新组** `system:cachemonitor:list` / `system:cachemonitor:delete` |
| 缓存名称 | **动态前缀聚合**（SCAN `gosilite:*`，known exact key 优先，否则截最后一段） |
| Hash | 中栏只 **真实 key**；field 不当 key；右栏 HGETALL JSON |

## Design

- 左：`prefix | count | remark`（remark 用 consts 已知前缀表）
- 中：`prefix` 下 keys + TYPE
- 右：String 原文 / Hash JSON；删单 key、按前缀清理（强确认）
- API：`/system/cache/{prefix/list,key/list,value,key/delete,prefix/clear}`
- 菜单 0017：`4703` C 缓存监控 + F 删除；system/admin 绑定
- CacheClient 扩展：`ScanAll` / `TypeOf`（有 limit，禁止 KEYS）

## Out of Scope

命令统计、集群多 DB、非 `gosilite:*` 键。

## Tasks

- [x] spec + IDL/hz + 0017
- [x] cache 扩展 + Service/Handler
- [x] 前端三栏 + check

## Report

**验收（2026-02-10）**

- `WEB=1 ./script/check.sh` 全绿；迁移 0017 已应用。
- 冒烟：`prefix/list` 返回 3 组真实前缀+备注；`key/list` 75 条 string；`value` 可读。
- 权限：`system:cachemonitor:list` / `system:cachemonitor:delete`。
