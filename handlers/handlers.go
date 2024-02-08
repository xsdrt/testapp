package handlers

import (
	"net/http"

	"github.com/xsdrt/hispeed2"
)

type Handlers struct {
	App *hispeed2.HiSpeed2
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "home", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error renderering:", err)
	}
}
