---
title: "Hugo × Hertz 集成"
---

## 核心思路

Hugo 和 Hertz 完全可以集成在一起，分工非常清晰：

- **Hugo** 负责生成静态文件（HTML/CSS/图片），输出到 `docs/hugo/public/`；
- **Hertz** 负责托管和提供这些文件，对外暴露 `/docs/` 路由。

文档站与后端同源部署，管理端「系统工具 → 开发文档」用 iframe 直接内嵌，无需独立的静态资源服务。

## Hertz 侧：adaptor + http.FileServer

Hertz 官方文档指出，原生 `app.FS` 实现维护状况不佳，**推荐通过适配器复用标准库的 `http.FileServer`**。本项目与 Swagger 的挂载方式一致（`main.go`）：

```go
import "github.com/cloudwego/hertz/pkg/common/adaptor"

// /docs → 301 到 /docs/
h.GET("/docs", func(c context.Context, ctx *app.RequestContext) {
    ctx.Redirect(http.StatusMovedPermanently, []byte("/docs/"))
})

// 静态文件托管（StripPrefix 后交给标准库 FileServer）
h.GET("/docs/*any", adaptor.HertzHandler(
    http.StripPrefix("/docs", http.FileServer(http.Dir(docsDir))),
))
```

要点：

| 项 | 约定 |
|----|------|
| 根目录 | conf `docs.dir`，默认 `docs/hugo/public`，可用环境变量 `GOSILITE_DOCS_DIR` 覆盖 |
| 鉴权 | 与 `/swagger` 一致，公开静态资源；管理端入口由菜单绑定控制 |
| 日志 | `/docs` 前缀加入 operlog 跳过名单，避免静态资源刷操作日志 |
| 目录不存在 | 标准库 FileServer 返回 404，不做特殊兜底 |

## Hugo 侧：构建链

```bash
script/bootstrap-hugo.sh   # 一次性：下载 pinned Hugo 到 script/bin/
script/build-docs.sh       # hugo → docs/hugo/public/（无 hugo 时 SKIP，exit 0）
```

- 站点源码在 `docs/hugo/`：`hugo.toml` + `layouts/`（自写极简模板）+ `content/`，不引主题、不联网，构建结果确定。
- 产物 `docs/hugo/public/` 已 gitignore，不入库。
- `./script/check.sh` 会自动调用 `build-docs.sh`：机器没装 hugo 打印 SKIP，装了但构建失败则让 check 变红。

## 访问路径

| 入口 | 地址 |
|------|------|
| 管理端菜单 | 系统工具 → 开发文档（iframe → `/docs/`） |
| 后端直链 | `http://127.0.0.1:8888/docs/` |
| 本地开发 | Vite 把 `/docs` 代理到 8888（`web/vite.config.ts`） |
