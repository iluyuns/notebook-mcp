package web

import (
	"net/http"

	"notebook-mcp/internal/httpapi"
	"notebook-mcp/internal/oauth"
	"notebook-mcp/internal/repo"
	"notebook-mcp/internal/service"

	"github.com/gin-gonic/gin"
)

// Deps 注册 HTTP 路由所需依赖；HTML 模板在 Register 内统一加载并注入 OAuth 页面。
type Deps struct {
	NoteSvc    *service.NoteService
	UserRepo   *repo.UserRepo
	OAuthSvc   *oauth.Service
	MCPHandler gin.HandlerFunc
	MCPPath    string
}

// Register 挂载首页、健康检查、REST API、OAuth、MCP。
func Register(engine *gin.Engine, d Deps) {
	tmpl := LoadPageTemplates()
	oauthH := oauth.NewHandler(d.OAuthSvc, d.UserRepo, oauth.HTMLTemplates{
		Authorize: tmpl.OAuthAuthorize,
		Register:  tmpl.OAuthRegister,
	})

	engine.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Home.Execute(c.Writer, nil)
	})
	engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := engine.Group("/api/v1")
	api.Use(oauth.BearerAuthMiddleware(d.OAuthSvc))
	httpapi.NewHandler(d.NoteSvc).Register(api)
	oauthH.Register(engine)

	mcpRoute := engine.Group(d.MCPPath)
	mcpRoute.Use(oauth.BearerAuthMiddleware(d.OAuthSvc))
	mcpRoute.Any("", d.MCPHandler)
}
