package handler

import (
	"net/http"

	"github.com/Livingpool/tim-tris/web"
)

type HttpHandler struct {
	renderer web.TemplatesInterface
}

func NewHttpHandler(renderer web.TemplatesInterface) *HttpHandler {
	return &HttpHandler{
		renderer: renderer,
	}
}

func (h *HttpHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, "home", nil)
}

func (h *HttpHandler) Chat(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, "chat", nil)
}
