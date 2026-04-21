package oauth

import (
	"html/template"
	"net/http"
	"net/url"

	"notebook-mcp/internal/repo"

	"github.com/gin-gonic/gin"
)

// HTMLTemplates 由 internal/web 统一加载后注入（authorize / register 页面）。
type HTMLTemplates struct {
	Authorize *template.Template
	Register  *template.Template
}

type Handler struct {
	svc   *Service
	users *repo.UserRepo
	pages HTMLTemplates
}

type authorizeView struct {
	ClientID            string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	State               string
	Error               string
}

type registerView struct {
	Success bool
	Error   string
}

func NewHandler(svc *Service, users *repo.UserRepo, pages HTMLTemplates) *Handler {
	return &Handler{svc: svc, users: users, pages: pages}
}

// Register 注册 OAuth 与注册页路由（不含首页；由 internal/web 统一编排）。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/.well-known/oauth-authorization-server", h.metadata)
	r.GET("/oauth/authorize", h.authorizePage)
	r.POST("/oauth/authorize", h.authorize)
	r.POST("/oauth/token", h.token)
	r.GET("/register", h.registerPage)
	r.POST("/register", h.register)
}

func (h *Handler) metadata(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Metadata())
}

func (h *Handler) authorizePage(c *gin.Context) {
	data, ok := h.authorizeQuery(c)
	if !ok {
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = h.pages.Authorize.Execute(c.Writer, data)
}

func (h *Handler) authorizeQuery(c *gin.Context) (authorizeView, bool) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	codeChallenge := c.Query("code_challenge")
	method := c.DefaultQuery("code_challenge_method", "S256")
	state := c.Query("state")

	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		c.String(http.StatusBadRequest, "missing required oauth query params")
		return authorizeView{}, false
	}
	return authorizeView{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: method,
		State:               state,
	}, true
}

func (h *Handler) authorize(c *gin.Context) {
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	codeChallenge := c.PostForm("code_challenge")
	method := c.DefaultPostForm("code_challenge_method", "S256")
	state := c.PostForm("state")
	username := c.PostForm("username")
	password := c.PostForm("password")

	view := authorizeView{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: method,
		State:               state,
	}

	userID, ok := h.svc.ValidateLogin(c.Request.Context(), username, password)
	if !ok {
		c.Header("Content-Type", "text/html; charset=utf-8")
		view.Error = "用户名或密码错误"
		_ = h.pages.Authorize.Execute(c.Writer, view)
		return
	}

	code, err := h.svc.IssueCode(userID, clientID, redirectURI, codeChallenge, method)
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		view.Error = err.Error()
		_ = h.pages.Authorize.Execute(c.Writer, view)
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

func (h *Handler) registerPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	v := registerView{Success: c.Query("ok") == "1"}
	_ = h.pages.Register.Execute(c.Writer, v)
}

func (h *Handler) register(c *gin.Context) {
	if h.users == nil {
		c.String(http.StatusServiceUnavailable, "registration unavailable")
		return
	}
	username := c.PostForm("username")
	password := c.PostForm("password")
	invite := c.PostForm("invite_code")

	_, err := h.users.RegisterWithInvite(c.Request.Context(), username, password, invite)
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		_ = h.pages.Register.Execute(c.Writer, registerView{Error: err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/register?ok=1")
}
