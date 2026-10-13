---
title: "发布 v0.1.0"
weight: 7
---

v0.1.0 是 gosilite 的首个正式基线。它使用一个最终数据库 schema 初始化新环境，不提供历史开发数据库的就地升级。发布必须从 `develop` 经 Pull Request 合并到 `main`；不要强推或直接改写 `main` 历史。

## 发布前提

- `WEB=1 ./script/check.sh` 通过。
- 根级 CI 通过，包括 PostgreSQL 16 的最终 schema 初始化。
- 生产环境已提供 `GOSILITE_JWT_SECRET`、`GOSILITE_DATABASE_DSN`、`GOSILITE_STORAGE_ACCESS_KEY` 和 `GOSILITE_STORAGE_SECRET_KEY`；值不写入仓库或 `.env.prod`。
- 已确认 PostgreSQL、Redis、对象存储的备份、网络访问与恢复负责人。

## 初始化数据库

最终 schema 位于 `script/schema/v0.1.0.sql`。初始化脚本只接受空数据库；发现 `system_user` 等核心表时会立即退出，绝不删除或覆盖已有对象。

本地 Docker 默认库：

```bash
./script/init-schema.sh
```

外部 PostgreSQL：

```bash
export GOSILITE_SCHEMA_DSN='postgresql://<user>:<password>@<host>:5432/<database>?sslmode=require'
./script/init-schema.sh

# DSN 模式要求部署机已安装 PostgreSQL 客户端（psql）。
```

已有开发库或生产前环境不支持就地升级。先完成备份、创建一个空数据库，再执行初始化。v0.1.0 不创建 `schema_migrations` 表。

## 生产启动与暴露面

```bash
export GO_ENV=prod
export GOSILITE_JWT_SECRET='<至少32字符的随机密钥>'
export GOSILITE_DATABASE_DSN='host=... user=... password=... dbname=... sslmode=require'
export GOSILITE_STORAGE_ACCESS_KEY='...'
export GOSILITE_STORAGE_SECRET_KEY='...'
go run .
```

生产环境不会注册 pprof，即使 `conf.yaml` 的本地开发默认值开启。将业务服务、Swagger、Hugo 文档和 Prometheus 指标置于反向代理或网络策略之后：

| 入口 | 建议 |
|---|---|
| `/ping` | 仅负载均衡与监控探活 |
| `/swagger/`、`/docs/` | 内部网络或管理员访问 |
| `:9091/metrics` | 仅监控系统访问，不对公网暴露 |
| pprof | 仅非生产调试环境 |

## 合并与发布顺序

1. 冻结 `develop`，只接收发布阻断修复。
2. 在干净 PostgreSQL 16 环境执行 `./script/init-schema.sh`，启动应用并检查 `GET /ping`。
3. 执行登录、菜单、权限、文件上传和定时任务日志的核心烟测。
4. 创建 `develop` 到 `main` 的 Pull Request，等待 CI 和审阅通过。
5. 正常合并 PR；不要 `push --force`，不要在未审阅状态下直接提交到 `main`。
6. 在 `main` 上创建并推送 `v0.1.0` 标签，发布说明包含数据库基线、运行环境和已知边界。
7. 生产部署顺序为：备份 → 创建空库 → 初始化 schema → 部署应用 → 探活与烟测。

应用二进制可以回退；数据库回退依赖部署前备份，因此在初始化前必须验证备份可恢复。
