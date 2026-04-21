# notebook-mcp

私有 HTTP MCP 服务，支持保存与查询个人上下文笔记（summary/context/note）。

## 功能

- MCP tools
  - `save_session_summary`
  - `save_session_context`
  - `save_session_note`
  - `save_by_instruction`（识别“总结/上下文/笔记”关键词）
  - `query_notes`
- HTTP API
  - `POST /api/v1/notes`
  - `GET /api/v1/notes/search`
- 首页（介绍与入口）
  - `GET /`
- Health
  - `GET /health`
- OAuth
  - `GET /.well-known/oauth-authorization-server`
  - `GET/POST /oauth/authorize`
  - `POST /oauth/token`
- 注册（需邀请码）
  - `GET/POST /register`

## 环境变量

- `POSTGRES_DSN`：PostgreSQL 连接串（必填）
- `PORT`：服务端口，默认 `8088`
- `MCP_PATH`：MCP 路径，默认 `/mcp`
- `DEFAULT_QUERY_LIMIT`：默认查询条数，默认 `20`
- `OAUTH_ISSUER_URL`：OAuth issuer URL，默认 `http://localhost:8088`
- `OAUTH_CLIENT_ID`：OAuth client id，默认 `cursor-private-notebook`
- `INITIAL_INVITE_CODE`：可选，启动时若设置则向 `notebook_invitation_codes` 插入一条可用邀请码（便于首次注册）
- `INITIAL_INVITE_MAX_USES`：与上一项配合，该码最大使用次数，默认 `1000`
- `OAUTH_CODE_TTL_SECONDS`：授权码 TTL，默认 `300`
- `OAUTH_ACCESS_TOKEN_TTL_SECONDS`：access token TTL，默认 `3600`

## 数据表

迁移：在 `docs/notebook` 下可直接 `make mu` / `make mv`（默认连接与 `docker-compose`、`.env.example` 一致，亦可用 `DATABASE_URL` 或 `POSTGRES_DSN` 覆盖）。跑服务仍需设置 `POSTGRES_DSN`。注册前需在 `notebook_invitation_codes` 中有可用邀请码，或使用 `INITIAL_INVITE_CODE` 自动插入。

## 本地 PostgreSQL（Docker）

`docker-compose.yml` 使用官方 `postgres:latest`。可复制 `.env.example` 为 `.env` 后执行 `docker compose up -d`。应用与迁移示例：`POSTGRES_DSN=postgres://notebook:notebook@127.0.0.1:5432/notebook?sslmode=disable`（与 `.env` 中账号库名一致）。

## 启动

```bash
cd docs/notebook
go run ./cmd/server
```

## Cursor MCP 配置

```json
{
  "mcpServers": {
    "private-notebook": {
      "url": "http://localhost:8088/mcp",
      "transport": "sse"
    }
  }
}
```

`/mcp` 现在需要 Bearer Token，需先走 OAuth 授权并换取 access token。

## 代码布局

- `internal/web`：HTML 模板目录 `templates/`，`web.Register` 统一挂载 `/`、健康检查、`httpapi`、`oauth`、MCP 路由。
- `internal/httpapi`：JSON REST。
- `internal/oauth`：OAuth 与注册表单处理逻辑；页面模板由 `web` 解析后通过 `oauth.HTMLTemplates` 注入。
