package oauth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/.well-known/oauth-authorization-server", h.metadata)
	r.GET("/oauth/authorize", h.authorizePage)
	r.POST("/oauth/authorize", h.authorize)
	r.POST("/oauth/token", h.token)
}

func (h *Handler) metadata(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Metadata())
}

func (h *Handler) authorizePage(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	codeChallenge := c.Query("code_challenge")
	method := c.DefaultQuery("code_challenge_method", "S256")
	state := c.Query("state")

	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(redirectURI) == "" || strings.TrimSpace(codeChallenge) == "" {
		c.String(http.StatusBadRequest, "missing required oauth query params")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<html><body><h3>Notebook MCP OAuth</h3><form method="post" action="/oauth/authorize">
<input type="hidden" name="client_id" value="`+htmlEscape(clientID)+`" />
<input type="hidden" name="redirect_uri" value="`+htmlEscape(redirectURI)+`" />
<input type="hidden" name="code_challenge" value="`+htmlEscape(codeChallenge)+`" />
<input type="hidden" name="code_challenge_method" value="`+htmlEscape(method)+`" />
<input type="hidden" name="state" value="`+htmlEscape(state)+`" />
<div>username: <input name="username" /></div>
<div>password: <input type="password" name="password" /></div>
<button type="submit">Authorize</button>
</form></body></html>`)
}

func (h *Handler) authorize(c *gin.Context) {
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	codeChallenge := c.PostForm("code_challenge")
	method := c.DefaultPostForm("code_challenge_method", "S256")
	state := c.PostForm("state")
	username := c.PostForm("username")
	password := c.PostForm("password")

	if !h.svc.ValidateLogin(username, password) {
		c.String(http.StatusUnauthorized, "invalid credentials")
		return
	}
	code, err := h.svc.IssueCode(clientID, redirectURI, codeChallenge, method)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	redirect, err := url.Parse(redirectURI)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid redirect_uri")
		return
	}
	q := redirect.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirect.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirect.String())
}

func (h *Handler) token(c *gin.Context) {
	if c.PostForm("grant_type") != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}
	token, expiresIn, err := h.svc.ExchangeCode(
		c.PostForm("code"),
		c.PostForm("code_verifier"),
		c.PostForm("client_id"),
		c.PostForm("redirect_uri"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"scope":        "mcp",
	})
}

func htmlEscape(v string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		`"`, "&quot;",
		"<", "&lt;",
		">", "&gt;",
		"'", "&#39;",
	)
	return replacer.Replace(v)
}
