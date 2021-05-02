package handler

import "github.com/urfave/cli/v2"

type Handler interface {
	// downloader
	Health() *cli.Command
	DownloadSong() *cli.Command
	DownloadPlaylist() *cli.Command
	DownloadAlbum() *cli.Command
	DownloadShow() *cli.Command
	SyncPlaylist() *cli.Command

	// download state alterations
	PausePlaylistDownload() *cli.Command
	ResumePlaylistDownload() *cli.Command

	// download progress trackers
	ViewSongProgress() *cli.Command
	ViewPlaylistProgress() *cli.Command
	ViewAlbumProgress() *cli.Command
	ViewShowProgress() *cli.Command

	// informational endpoints
	ViewPlaylistMeta() *cli.Command
	ViewAlbumMeta() *cli.Command
	ViewSongMeta() *cli.Command
	ViewShowMeta() *cli.Command

	// player
	PlayPauseSong() *cli.Command
}
