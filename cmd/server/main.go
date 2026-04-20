package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"notebook-mcp/internal/config"
	"notebook-mcp/internal/db"
	"notebook-mcp/internal/httpapi"
	"notebook-mcp/internal/mcpserver"
	"notebook-mcp/internal/oauth"
	"notebook-mcp/internal/repo"
	"notebook-mcp/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pg, err := db.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("connect postgres failed: %v", err)
	}
	defer pg.Close()

	noteRepo := repo.NewNoteRepo(pg)
	noteSvc := service.NewNoteService(noteRepo, cfg.DefaultPerPage)
	mcpHandler := mcpserver.New(noteSvc)
	oauthSvc := oauth.NewService(oauth.Config{
		IssuerURL:      cfg.OAuthIssuerURL,
		ClientID:       cfg.OAuthClientID,
		User:           cfg.OAuthUser,
		Password:       cfg.OAuthPassword,
		CodeTTL:        time.Duration(cfg.CodeTTL) * time.Second,
		AccessTokenTTL: time.Duration(cfg.AccessTokenTTL) * time.Second,
	})

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	httpapi.NewHandler(noteSvc).Register(r)
	oauth.NewHandler(oauthSvc).Register(r)
	mcpRoute := r.Group(cfg.MCPPath)
	mcpRoute.Use(oauth.BearerAuthMiddleware(oauthSvc))
	mcpRoute.Any("", gin.WrapH(mcpHandler))

	addr := ":" + cfg.Port
	log.Printf("notebook mcp service listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
