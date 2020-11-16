package handler

import (
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/middleware"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/core"
	"github.com/gorilla/mux"
)

func RegisterHandler(r *mux.Router, svc core.Service) {
	coreHandler := NewHandler(r, svc)
	middleware.RegisterMiddlewares(r)
	r.Handle("/ping", coreHandler.Health())
}
