# Hertz / hz API 索引（gosilite 实用子集）

官方文档：https://cloudwego.cn/zh/docs/hertz/  
下列条目已与本仓库用法对齐；不确定的 API **先查官方文档**，不要凭 Gin 记忆编写。

## Handler

```go
func(ctx context.Context, c *app.RequestContext) {
    c.JSON(consts.StatusOK, utils.H{"message": "pong"})
}
```

- 第一参数是 `context.Context`，第二参数是 `*app.RequestContext`。
- 需要写业务响应时统一用 `pkg/utils/response`。

## 路由

```go
h.GET("/ping", handler)
h.POST("/user/login", handler)
g := r.Group("/user")
g.GET("/info", append(mw, user.GetInfo)...)
```

常用方法：`GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS/Handle/Any/Group/Use`。

## 绑定与校验

```go
var req user.LoginRequest
if err := c.BindAndValidate(&req); err != nil { ... }
```

| API | 用途 |
|-----|------|
| `BindAndValidate` | 绑定 + 校验（常用） |
| `Bind` | 只绑定 |
| `BindQuery` / `BindHeader` / `BindPath` / `BindForm` / `BindJSON` | 单来源 |
| `BindByContentType` | 按 Content-Type |
| `Validate` | 只校验 |

绑定优先级（官方）：`path > form > query > cookie > header > json > raw_body`。

常用 tag：`path` `form` `query` `cookie` `header` `json` `raw_body` `default` `required` `validate`。

## 中间件

```go
type HandlerFunc func(ctx context.Context, c *app.RequestContext)

h.Use(func(ctx context.Context, c *app.RequestContext) {
    c.Next(ctx)
})
```

本项目入口顺序见 `docs/patterns/design.md` §6。

## Engine

- `server.Default()`：默认带 Recovery。
- `server.New(server.WithHostPorts(addr))`：完全控制（本项目 `main.go` 使用），须自装 Recovery。
- `h.Spin()`：推荐启动方式，支持优雅退出。
- `Run()` 不提供默认优雅退出。

配置项：`WithHostPorts`、Read/Write/IdleTimeout、MaxRequestBodySize、KeepAlive、TLS、H2C 等。

## hz 代码生成

```bash
go install github.com/cloudwego/hertz/cmd/hz@latest

hz update -I idl -idl idl/framework/user.thrift -customize_package idl/package.yaml
```

- `.hz` 记录生成目录：`api/handler`、`api/model`、`api/router`。
- IDL 路由注解：`(api.get="/user/info")`、`(api.post="/user/login")`。
- 生成文件勿手改；`middleware.go` 仅改权限挂载。

## 本项目响应封装

```go
R.SuccessWithData(ctx, c, data)     // 200 {code:200,msg:"success",data}
R.Error(ctx, c, err)                // 业务码或 500
R.Unauthorized(c, msg)              // 401
R.InvalidParamRequestResponse(c, m) // 400
```

## JWT 会话

- 中间件：`github.com/hertz-contrib/jwt`
- 登录回调 `LoginResponse` 输出 `SessionTokenResponse`（camelCase）
- 受保护路由：`Authorization: Bearer <accessToken>`
- 刷新是**独立业务路由** `/user/refresh`（refresh token 在 Redis），不是 JWT 库默认 `RefreshHandler` 的 token 延长。

## 启动

```bash
export PATH="$HOME/.local/go/bin:$PATH"
GO_ENV=test go run .
# 监听 conf/test/conf.yaml 的 server.address，默认 :8888
```
