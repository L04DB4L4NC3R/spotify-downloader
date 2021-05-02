package handler

import (
	"encoding/json"
	"net/http"

	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/views"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/pkg/core"
	"github.com/gorilla/mux"
)

type handler struct {
	router  *mux.Router
	service core.Service
}

func (h *handler) ViewPlaylistMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistMeta, err := h.service.FetchPlaylistMeta(core.RESOURCE_PLAYLIST, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Playlist Metadata", playlistMeta, http.StatusOK)
	})
}

func (h *handler) ViewAlbumMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistMeta, err := h.service.FetchPlaylistMeta(core.RESOURCE_ALBUM, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Album Metadata", playlistMeta, http.StatusOK)
	})
}

func (h *handler) ViewShowMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistMeta, err := h.service.FetchPlaylistMeta(core.RESOURCE_SHOW, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Show Metadata", playlistMeta, http.StatusOK)
	})
}

func (h *handler) ViewSongMeta() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		songMeta, err := h.service.FetchSongMeta(vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Song Metadata", songMeta, http.StatusAccepted)
	})
}

func (h *handler) ViewSongProgress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		status, err := h.service.CheckSongStatus(vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Song Progress", status, http.StatusOK)
	})
}

func (h *handler) ViewPlaylistProgress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		status, err := h.service.CheckPlaylistStatus(core.RESOURCE_PLAYLIST, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Playlist Progress", status, http.StatusOK)
	})
}

func (h *handler) ViewAlbumProgress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		status, err := h.service.CheckPlaylistStatus(core.RESOURCE_ALBUM, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Album Progress", status, http.StatusOK)
	})
}

func (h *handler) ViewShowProgress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		status, err := h.service.CheckPlaylistStatus(core.RESOURCE_SHOW, vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Show Progress", status, http.StatusOK)
	})
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
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_PLAYLIST, vars["id"], nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Playlist Metadata", playlistmetas, http.StatusAccepted)
	})
}

func (h *handler) DownloadAlbum() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_ALBUM, vars["id"], nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Album Metadata", playlistmetas, http.StatusAccepted)
	})
}

func (h *handler) DownloadShow() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_SHOW, vars["id"], nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err.Error(), http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Show Metadata", playlistmetas, http.StatusAccepted)
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

func (h *handler) FetchResourceSongs() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		meta, err := h.service.FetchPlaylistSongMetas(vars["resource"], vars["id"])
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Resource Song Meta", meta, http.StatusOK)
	})
}

func (h *handler) ViewBulkSongProgress() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var songs struct {
			SongIds []string `json:"song_ids"`
		}
		json.NewDecoder(r.Body).Decode(&songs)
		meta, err := h.service.CheckBulkSongStatus(songs.SongIds)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		views.Fill(w, "Bulk Song Status", meta, http.StatusOK)
	})
}
func NewHandler(r *mux.Router, svc core.Service) Handler {
	return &handler{
		router:  r,
		service: svc,
	}
}
