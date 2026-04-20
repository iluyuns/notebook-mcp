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
- Health
  - `GET /health`
- OAuth
  - `GET /.well-known/oauth-authorization-server`
  - `GET/POST /oauth/authorize`
  - `POST /oauth/token`

## 环境变量

- `POSTGRES_DSN`：PostgreSQL 连接串（必填）
- `PORT`：服务端口，默认 `8088`
- `MCP_PATH`：MCP 路径，默认 `/mcp`
- `DEFAULT_QUERY_LIMIT`：默认查询条数，默认 `20`
- `OAUTH_ISSUER_URL`：OAuth issuer URL，默认 `http://localhost:8088`
- `OAUTH_CLIENT_ID`：OAuth client id，默认 `cursor-private-notebook`
- `OAUTH_USER`：授权页登录用户名（必填）
- `OAUTH_PASSWORD`：授权页登录密码（必填）
- `OAUTH_CODE_TTL_SECONDS`：授权码 TTL，默认 `300`
- `OAUTH_ACCESS_TOKEN_TTL_SECONDS`：access token TTL，默认 `3600`

## 数据表

执行 `migrations/000001_create_notes.up.sql` 初始化表结构。

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
