package handler

import (
	"net/http"
)

type Handler interface {
	// downloader
	Health() http.Handler
	DownloadSong() http.Handler
	// DownloadPlaylist() http.Handler
	// SyncPlaylist() http.Handler

	// // download state alterations
	// PausePlaylistDownload() http.Handler
	// ResumePlaylistDownload() http.Handler

	// // download progress trackers
	// ViewPlaylistMeta() http.Handler
	// ViewSongMeta() http.Handler
	// ViewProgressOfPlaylistDownload() http.Handler

	// // player
	// PlayPauseSong() http.Handler
}
