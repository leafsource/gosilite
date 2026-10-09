# Feature Spec: 系统监控 · 系统接口

- status: delivered
- created: 2026-02-10

## Problem

联调/运维需要在管理端直接查阅全部 API。Swagger 已挂载于后端 `/swagger`，菜单层提供入口即可。

## 拍板

| 决策点 | 拍板 |
|--------|------|
| 形式 | **内嵌 Swagger UI**（iframe），不自研接口文档引擎 |
| 路径 | 经 Vite/API 前缀代理：`{VITE_API_URL_PREFIX}/swagger/index.html` |
| 权限 | `system:apimonitor:list` |
| 菜单 | 4704 C 系统接口 → `monitor/api/index` |

## Design

- 页面：全高 iframe 指向 Swagger UI；加载失败给出提示与后端直链。
- 不改 Swagger 注解范围；注解不全属文档质量问题，不阻塞本页。

## Out of Scope

在线调试增强、按 IDL 自动生成 OpenAPI、权限码自动映射表。

## Tasks

- [x] 菜单迁移 0018
- [x] 前端页 + 路由
- [x] check

## Report

验收：菜单可见；iframe 打开 Swagger；`WEB=1 ./script/check.sh` 绿。
