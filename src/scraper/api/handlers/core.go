package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/views"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/pkg/core"
	"github.com/gorilla/mux"
)

var (
	errInvalidInput = errors.New("invalid input")
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, format, err := h.parseIdFormat(vars["id"])
		if err != nil {
			views.Fill(w, "bad input", err, http.StatusBadRequest)
			return
		}
		playlistmetas, err := h.service.PlaylistSync(core.RESOURCE_PLAYLIST, id, format, nil)
		if err != nil {
			views.Fill(w, "Some error occurred", err, http.StatusInternalServerError)
			return
		}
		// Fireforget
		// gRPC transport is handled in the service
		views.Fill(w, "Playlist Metadata", playlistmetas, http.StatusAccepted)
	})
}

func (h *handler) DownloadPlaylist() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, format, err := h.parseIdFormat(vars["id"])
		if err != nil {
			views.Fill(w, "bad input", err, http.StatusBadRequest)
			return
		}
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_PLAYLIST, id, format, nil)
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
		id, format, err := h.parseIdFormat(vars["id"])
		if err != nil {
			views.Fill(w, "bad input", err, http.StatusBadRequest)
			return
		}
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_ALBUM, id, format, nil)
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
		id, format, err := h.parseIdFormat(vars["id"])
		if err != nil {
			views.Fill(w, "bad input", err, http.StatusBadRequest)
			return
		}
		playlistmetas, err := h.service.PlaylistDownload(core.RESOURCE_SHOW, id, format, nil)
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
		var (
			vars = mux.Vars(r)
		)
		id, format, err := h.parseIdFormat(vars["id"])
		if err != nil {
			views.Fill(w, "invalid requst", err, http.StatusBadRequest)
			return
		}
		songMeta, err := h.service.SongDownload(id, format, nil)
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

func (h *handler) parseIdFormat(payload string) (id, format string, err error) {
	idString := strings.Split(payload, ".")
	n := len(idString)
	switch n {
	case 0:
		return "", "", err
	case 1:
		id = idString[0]
		format = "mp3"
		return id, format, nil
	case 2:
		id = idString[0]
		switch idString[1] {
		case "mp3", "flac", "opus":
			format = idString[1]
			return id, format, nil
		default:
			return "", "", errInvalidInput
		}
	default:
		return "", "", errInvalidInput
	}
}
