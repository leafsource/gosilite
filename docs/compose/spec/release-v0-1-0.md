---
feature: release-v0-1-0
status: in-progress
updated: 2026-10-12
branch: release/v0.1.0
commits:
---

# v0.1.0 发布基线

## Report

## [S1] Problem

`develop` 已具备完整业务能力，但 `main` 仍是初始提交。数据库同时维护一份过时的 Navicat 基线和 29 个历史增量迁移；仓库没有根级 CI，发布文档也未覆盖调度器、最终数据库基线和新增业务域的标准路径。直接合并会让首次正式发布不可重复、难以审计。

## [S2] Design

### S2.1 发布数据库基线

v0.1.0 采用“仅最终基线”策略，不兼容此前开发库的增量升级，也不引入 `schema_migrations` 版本表。发布前须备份旧库；新环境只执行 `script/schema/v0.1.0.sql`。

该 schema 文件必须是 PostgreSQL 16 的完整初始化入口，包含当前最终 DDL、索引/约束以及最小系统种子：唯一 `skyx` 系统管理员、`system` 角色、完整菜单/权限、基础字典、参数和本地对象存储配置；不得包含演示账号、运行日志、消息记录或调度运行记录。历史 `script/postgres.sql` 与整个 `script/migrations/` 目录从该发布分支移除。提供一个轻量脚本，用目标数据库 DSN（需要 psql 客户端）或 Docker 本地数据库执行该 schema；脚本只接受空数据库，发现核心表时拒绝执行，且绝不自动删除数据库或表。发布文档必须明确：已有旧库需要先备份再重建，v0.1.0 不提供就地升级。

### S2.2 工程化发布门槛

增加根级 GitHub Actions 工作流，在 PR 与 `main` 推送时固定 Go、Node、PostgreSQL 16、Redis 7 环境，执行最终 schema 初始化和 `WEB=1 ./script/check.sh`。Hugo 成为 CI 中的实际依赖，文档构建不得因缺少 Hugo 被跳过。

新增发布清单与版本说明：生产配置要求、数据库初始化顺序、部署后探活、核心冒烟、回滚边界（应用可回退，数据按备份恢复）。`GO_ENV=prod` 默认不暴露 pprof；开发环境保持现有可选调试能力。

### S2.3 文档与 Agent 开发路径

同步 README、Hugo 文档、AGENTS 与技能规则，使其准确反映 v0.1.0：最终 schema、调度器、生产访问面、CI 和发布步骤。

新增面向使用者的“新增业务模块”文档，固定以下顺序：需求/权限 → schema 变更 → gorm 生成 → IDL/hz 生成 → Repository → Service → Handler/Swagger → Module 装配 → 菜单与前端 → 验收。v0.1.0 的 schema 为单一最终快照；后续版本的数据库演进策略在实际发布时单独确定。

## [S3] Out of Scope

- 不提供旧开发数据库的在线/增量升级或自动数据迁移。
- 不引入重型迁移框架、微服务、容器编排或发布平台。
- 不修改既有业务 HTTP 契约或重构已交付领域。
- 不在本工作流中合并、推送或打 `main` 标签；完成后由用户选择收尾动作。

## Tasks

- [ ] T1: 建立 `v0.1.0` 最终 schema 与非破坏性初始化脚本，删除历史迁移入口 — acceptance: 空 PostgreSQL 16 库可成功初始化；已有核心表的数据库被拒绝且不创建 schema 版本记录表 (covers: S2.1)
- [ ] T2: 调整生产默认安全配置与根级 CI 发布门槛 — acceptance: prod 默认不注册 pprof；CI 在服务容器上初始化 schema 后通过完整检查 (covers: S2.2; depends: T1)
- [ ] T3: 同步发布、数据库、调度器和新增业务模块文档 — acceptance: 所有命令与目录/版本事实一致，公开文档包含模块创建及 schema 快照流程 (covers: S2.1, S2.2, S2.3; depends: T1, T2)
- [ ] T4: 在干净数据库完成初始化和应用冒烟，执行全量检查 — acceptance: 空库初始化成功、重复执行安全拒绝；应用能启动并通过 `/ping` 与管理员登录，`WEB=1 ./script/check.sh` 通过 (covers: S2.1, S2.2; depends: T1, T2, T3)
- [ ] T5: 对完整发布变更进行独立评审并完成交付记录 — acceptance: 评审确认规格一致性、正确性和代码库约定，无 critical 发现 (covers: S2.1, S2.2, S2.3; depends: T4)
