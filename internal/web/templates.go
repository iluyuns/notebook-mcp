package web

import (
	"embed"
	"html/template"
)

//go:embed templates
var templateRoot embed.FS

// PageTemplates 由 internal/web/templates 下全部 HTML 解析得到，供路由与 oauth 页面复用。
type PageTemplates struct {
	Home           *template.Template
	OAuthAuthorize *template.Template
	OAuthRegister  *template.Template
}

// LoadPageTemplates 解析 internal/web/templates 下全部页面模板。
func LoadPageTemplates() PageTemplates {
	return PageTemplates{
		Home:           template.Must(template.ParseFS(templateRoot, "templates/home/index.html")),
		OAuthAuthorize: template.Must(template.ParseFS(templateRoot, "templates/oauth/authorize.html")),
		OAuthRegister:  template.Must(template.ParseFS(templateRoot, "templates/oauth/register.html")),
	}
}
