package handler

import (
	"net/http"

	"github.com/L04DB4L4NC3R/spotify-downloader/src/scraper/api/views"
	"github.com/L04DB4L4NC3R/spotify-downloader/src/scraper/core"
	"github.com/gorilla/mux"
)

type handler struct {
	router  *mux.Router
	service core.Service
}

func (h *handler) ViewPlaylistMeta() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) ViewSongMeta() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) ViewProgressOfPlaylistDownload() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) PausePlaylistDownload() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) ResumePlaylistDownload() http.Handler {
	panic("not implemented") // TODO: Implement
}

// player
func (h *handler) PlayPauseSong() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) SyncPlaylist() http.Handler {
	panic("not implemented") // TODO: Implement
}

func (h *handler) DownloadPlaylist() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistmetas, err := h.service.PlaylistDownload(vars["id"], nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Playlist Metadata", playlistmetas, http.StatusAccepted)
	})
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
