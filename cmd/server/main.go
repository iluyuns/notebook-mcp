package main

import (
	"context"
	"log"
	"time"

	"notebook-mcp/internal/config"
	"notebook-mcp/internal/db"
	"notebook-mcp/internal/mcpserver"
	"notebook-mcp/internal/oauth"
	"notebook-mcp/internal/repo"
	"notebook-mcp/internal/service"
	"notebook-mcp/internal/web"

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
	userRepo := repo.NewUserRepo(pg)
	if cfg.InitialInviteCode != "" {
		if err := userRepo.EnsureInviteCode(ctx, cfg.InitialInviteCode, cfg.InitialInviteMaxUses); err != nil {
			log.Fatalf("ensure invite code: %v", err)
		}
	}
	noteSvc := service.NewNoteService(noteRepo, cfg.DefaultPerPage)
	mcpHandler := mcpserver.New(noteSvc)
	oauthSvc := oauth.NewService(oauth.Config{
		IssuerURL:      cfg.OAuthIssuerURL,
		ClientID:       cfg.OAuthClientID,
		CodeTTL:        time.Duration(cfg.CodeTTL) * time.Second,
		AccessTokenTTL: time.Duration(cfg.AccessTokenTTL) * time.Second,
	}, userRepo)

	r := gin.Default()
	web.Register(r, web.Deps{
		NoteSvc:    noteSvc,
		UserRepo:   userRepo,
		OAuthSvc:   oauthSvc,
		MCPHandler: gin.WrapH(mcpHandler),
		MCPPath:    cfg.MCPPath,
	})

	addr := ":" + cfg.Port
	log.Printf("notebook mcp service listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
