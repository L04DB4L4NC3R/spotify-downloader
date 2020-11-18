package handler

import (
	"net/http"

	"github.com/L04DB4L4NC3R/spotify-downloader/ytber/api/views"
	"github.com/L04DB4L4NC3R/spotify-downloader/ytber/core"
	"github.com/gorilla/mux"
)

type handler struct {
	router  *mux.Router
	service core.Service
}

func (h *handler) DownloadSong() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		songMeta, err := h.service.SongDownload(vars["id"], nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Song Metadata", songMeta, http.StatusAccepted)
	})
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
