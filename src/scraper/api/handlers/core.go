package handler

import (
	"net/http"

	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/views"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/core"
	"github.com/gorilla/mux"
)

type handler struct {
	router  *mux.Router
	service core.Service
}

func (h *handler) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		views.Fill(w, "pong", nil, http.StatusOK)
	})
}

func NewHandler(r *mux.Router, svc core.Service) Handler {
	return &handler{
		router:  r,
		service: svc,
	}
}
