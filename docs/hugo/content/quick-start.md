---
title: "快速开始"
weight: 2
---

本页带你从零把 gosilite 跑起来：Docker 起基础设施 → Go 起后端（`:8888`）→ Vite 起前端（`:3002`）→ 测试账号登录 → 跑一遍验收脚本 → （可选）构建开发文档站。命令均与仓库 `README.md`、`script/` 实测一致，直接整段复制即可。想先了解项目结构，先读[概览](/docs/overview/)。

## 1. 前置条件

| 依赖 | 要求 | 说明 |
|---|---|---|
| Docker | 含 `docker compose` | PostgreSQL / Redis / RustFS 全部容器化 |
| Go | `go.mod` 声明 `go 1.26.0`；README 记录本机为 go1.27.x | 装在 `$HOME/.local/go/bin`，`go` 不在 PATH 时需先导出 |
| Node | ≥ 22.12.0（`web/package.json` 的 `engines`） | 前端依赖安装与开发服务器 |
| Hugo | 可选 | 缺失时文档站构建打印 `SKIP`，不阻塞验收 |

`go` 命令找不到时先执行：

```bash
export PATH="$HOME/.local/go/bin:$PATH"   # 本机 Go 路径，见 README
```

## 2. 起基础设施

在仓库根目录：

```bash
docker compose up -d
```

| 服务 | 宿主机端口 | 用途 |
|---|---|---|
| PostgreSQL 16 | 5432 | 业务与会话数据（库/用户/密码均为 `gosilite`） |
| Redis 7 | 6379 | refresh token 轮换、denylist（吊销名单） |
| RustFS | 19000（S3 API）/ 19001（控制台 UI） | S3 兼容对象存储；host 9000 被占用，故映射到 19000 |

连接参数不用改：`conf/conf.yaml` 的本地默认值（DSN `127.0.0.1:5432`、Redis `127.0.0.1:6379`、存储 `http://127.0.0.1:19000`）与上表一一对应，密钥与环境差异的覆盖方式见 `.env.example`。

## 3. 初始化数据库并起后端

v0.1.0 只支持在空 PostgreSQL 数据库上初始化最终 schema；脚本会拒绝覆盖已有核心表。

```bash
./script/init-schema.sh

export PATH="$HOME/.local/go/bin:$PATH"   # go 不在 PATH 时
GO_ENV=test go run .                       # Hertz 监听 :8888
```

另开终端探活：

```bash
curl -s http://127.0.0.1:8888/ping        # 返回 {"ping":"pong"} 即就绪
```

- 入口是仓库根目录的 `main.go`，监听地址来自 `conf/conf.yaml` 的 `server.address`（`:8888`）。
- `GO_ENV`（`test|dev|prod`，默认 `test`）决定加载哪个 `.env.<GO_ENV>` 文件，模板见 `.env.example`；本地默认值都在 `conf/conf.yaml`，无需改动。

## 4. 起前端

```bash
cd web && npm install && npm run dev
```

开发服务器端口为 **3002**（`web/vite.config.ts` 的 `server.port`；README 里的 `:5173` 为旧值，以配置为准），并配置两条代理：

| 浏览器请求前缀 | 代理规则 | 目标 |
|---|---|---|
| `/api/*` | 剥掉 `/api` 前缀后转发（`VITE_API_URL_PREFIX` 来自 `web/.env*`，值为 `/api`） | `http://127.0.0.1:8888` |
| `/docs/*` | 路径原样透传 | `http://127.0.0.1:8888`（Hertz 托管的文档站） |

本地后端路由本身没有 `/api` 前缀，所以前端统一以 `/api/user/login` 这类路径访问，代理负责去掉前缀；`/docs` 不剥前缀，因为文档站就挂在后端的 `/docs/` 下。

## 5. 登录

浏览器打开 `http://127.0.0.1:3002`，用本地测试账号（dev/test 库，系统管理员，见 `AGENTS.md`）登录：

| 项 | 值 |
|---|---|
| 账号 | `skyx` |
| 密码 | `Admin@123` |

登录接口为 `POST /user/login`（经 `/api` 代理即 `/api/user/login`），响应信封是统一的 `{"code":200,"msg":"success","data":...}`，`data` 为会话字段——OAuth2 风格 **camelCase**，没有 `token`/`expire` 这类旧字段：

| 字段 | 含义 |
|---|---|
| `accessToken` | 访问令牌（JWT），后续请求放 `Authorization: Bearer <accessToken>` |
| `tokenType` | 令牌类型，固定 `Bearer` |
| `expiresIn` | 访问令牌剩余有效期，单位秒 |
| `refreshToken` | 刷新令牌，存于 Redis，过期后 `POST /user/refresh` 换新并轮换 |

其余会话接口：`POST /user/refresh`、`POST /user/logout`、`POST /user/password`、`GET /user/info`、`GET /user/profile`。401 / 403 与菜单问题见[常见问题](/docs/faq/)。

## 6. 验收

```bash
export PATH="$HOME/.local/go/bin:$PATH"   # check.sh 也会自行补 PATH
./script/check.sh        # go test + go vet + go build + 文档站构建
WEB=1 ./script/check.sh  # 额外执行 web 的 npm run build
```

- 全部通过会输出 `check OK`。
- 文档站构建缺 hugo 时打印 `SKIP` 并继续（`script/build-docs.sh` 的设计，不阻塞验收）。
- 提交前的完整验收以 `WEB=1 ./script/check.sh` 全绿为准。

## 7. 构建文档站（可选）

```bash
script/bootstrap-hugo.sh   # 一次性：下载/源码构建 pinned 版本 Hugo 到 script/bin/hugo（可重复执行）
script/build-docs.sh       # 构建 docs/hugo → docs/hugo/public
```

- 未安装 hugo 也没有 `script/bin/hugo` 时，`build-docs.sh` 打印 `SKIP` 并以 0 退出；`bootstrap-hugo.sh` 默认安装 `HUGO_VERSION=0.167.0`。
- 构建完成后访问 `http://127.0.0.1:8888/docs/`（或经前端代理 `http://127.0.0.1:3002/docs/`）即可看到本站，管理端"开发文档"页会以 iframe 内嵌同一内容。

## 下一步

- [概览](/docs/overview/) — 技术栈、架构依赖与能力清单
- [后端](/docs/backend/) — 分层职责、逐模块说明、生成边界
- [前端](/docs/frontend/) — 动态路由、组件与开发规范
- [新增业务模块](/docs/module-development/) — 从 schema、IDL 到前端与验收的固定路径
- [发布 v0.1.0](/docs/release/) — 数据库初始化、生产访问面与 main 合并步骤
- [常见问题](/docs/faq/) — 401/403、菜单不显示、schema 与生成物等踩坑速查
- [参考](/docs/reference/) — 命令、配置、信封与权限码速查
