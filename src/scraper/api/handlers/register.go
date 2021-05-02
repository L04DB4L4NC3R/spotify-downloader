package handler

import (
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/middleware"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/pkg/core"
	"github.com/gorilla/mux"
)

func RegisterHandler(r *mux.Router, svc core.Service) {
	coreHandler := NewHandler(r, svc)
	middleware.RegisterMiddlewares(r)

	r.Handle("/ping/", coreHandler.Health())

	// queue downloads
	r.Handle("/song/{id}/", coreHandler.DownloadSong())
	r.Handle("/playlist/{id}/", coreHandler.DownloadPlaylist())
	r.Handle("/album/{id}/", coreHandler.DownloadAlbum())
	//r.Handle("/show/{id}/", coreHandler.DownloadShow())

	// fetch just the metadata
	r.Handle("/meta/song/{id}/", coreHandler.ViewSongMeta())
	r.Handle("/meta/playlist/{id}/", coreHandler.ViewPlaylistMeta())
	r.Handle("/meta/album/{id}/", coreHandler.ViewAlbumMeta())
	r.Handle("/metas/{resource}/{id}/", coreHandler.FetchResourceSongs())
	//r.Handle("/meta/show/{id}/", coreHandler.ViewShowMeta())

	// status info
	r.Handle("/status/song/{id}/", coreHandler.ViewSongProgress())
	// POST ids
	r.Handle("/status/songs/", coreHandler.ViewBulkSongProgress())
	r.Handle("/status/playlist/{id}/", coreHandler.ViewPlaylistProgress())
	r.Handle("/status/album/{id}/", coreHandler.ViewAlbumProgress())
	//r.Handle("/status/show/{id}/", coreHandler.ViewShowProgress())
}
