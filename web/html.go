package web

import (
	"embed"
	"html/template"
	"io"
)

//go:embed css/* html/* scripts/*
var StaticFiles embed.FS

type TemplatesInterface interface {
	Render(w io.Writer, name string, data interface{}) error
}

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("./web/html/*.html")),
	}
}
