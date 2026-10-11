---
title: "开发指南"
weight: 2
---

## 本地启动

```bash
docker compose up -d          # postgres / redis / rustfs
GO_ENV=test go run .          # 后端 :8888
cd web && npm run dev         # 前端（默认 :3002，本机常用 3003）
```

本地测试账号（dev/test 库，系统管理员）：`skyx` / `Admin@123`。

## 验收回路

```bash
./script/check.sh             # go test + vet + build
WEB=1 ./script/check.sh       # 附带前端 build
```

`check.sh` 末尾会执行 `script/build-docs.sh`（缺 hugo 时 SKIP，不影响全绿）。

## 关键约定速查

| 事项 | 约定 |
|------|------|
| HTTP 契约 | **先改 `idl/*.thrift` → `hz update` → 再改 Handler**；路由必须由 hz 从 IDL 生成 |
| 表结构 | `script/migrations/NNNN_*.sql` 幂等 DDL + 种子 → gorm gen |
| 分层 | Handler 绑定/校验/回包 → Service（具体类型，无 I*Service）→ Repository |
| 响应 | 只经 `pkg/utils/response`，信封 `{code,msg,data}` |
| 前端路由 | 侧栏与 path 唯一来源是 `GET /menu/routes`（数据库菜单）；页面 `defineOptions({name})` 必须等于 `genRouteName(fullPath)` |
| 权限 | 菜单 perms 控制 API；按钮 F 行挂权限码；角色-菜单绑定在 `system_role_menu` |
| Swagger | 每个 HTTP Handler 必须写完整 swag 注解；`GET /swagger/index.html` 查看 |
| 组件 | ≥3 处复用抽 `web/src/components/s-<name>/`；`t-*` API 先查 TDesign MCP |

## 目录结构（后端）

```text
main.go                 # 唯一生产入口（含 /swagger、/docs 挂载）
api/handler/            # Handler：绑定、校验、调 Service、回包
api/router/             # 路由注册（hz 生成骨架）
api/middleware/         # JWT、权限、ctx cache、operlog
application/<域>/       # Module 插件 + service + dal
conf/                   # conf.yaml 单一配置，密钥走 .env.<GO_ENV>
idl/                    # Thrift 契约源（hz 的输入）
script/migrations/      # 幂等迁移 SQL
web/                    # TDesign Vue3 前端
docs/hugo/              # 本 Hugo 文档站
```
