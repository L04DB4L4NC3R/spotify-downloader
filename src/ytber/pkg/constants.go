package pkg

import (
	"runtime"
	"time"
)

const (
	STATUS_META_FED         = "FED"
	STATUS_DWN_QUEUED       = "QUEUED"
	STATUS_DWN_FAILED       = "FAILED"
	STATUS_DWN_COMPLETE     = "COMPLETED"
	STATUS_THUMBNAIL_FAILED = "THUMBNAIL_FAILED"
	STATUS_FINISHED         = "FINISHED"

	YT_BASE_URL = "https://youtube.com/watch?v="

	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""

	YT_DOWNLOAD_METADATA_ARGS = " --add-metadata --postprocessor-args $'-metadata artist=\"%s\" -metadata title=\"%s\" -metadata date=\"%s\" -metadata purl=\"%s\" -metadata track=\"%s\"'"

	YT_DOWNLOAD_PATH_CMD = " -o \"music/%(title)s.%(ext)s\""

	// image url, song path, download path, title, song path
	FFMPEG_THUMBNAIL_CMD = "ffmpeg -y -i %s -i \"%s\" -map_metadata 1 -map 1 -map 0 \"%s/%s -(%s)-(%s).mp3\" && rm \"%s\""

	RESOURCE_PLAYLIST = "playlists"
	RESOURCE_SONG     = "tracks"
	RESOURCE_ALBUM    = "albums"

	// after this, song download is cancelled
	SONG_DOWNLOAD_TIMEOUT = time.Duration(2) * time.Minute

	// after this, remaining songs in batch are cancelled and queued for retry
	BATCH_DOWNLOAD_TIMEOUT = time.Duration(5) * time.Minute

	// after this,thumbnail application is cancelled
	// larger timeout due to song album art mapping
	THUMBNAIL_APPLICATION_TIMEOUT = time.Duration(30) * time.Minute
	THUMBNAIL_TEARDOWN_TIMEOUT    = time.Duration(300) * time.Minute

	// wait this much time before attempting to check whether song is downloaded
	SONG_DOWNLOAD_WAIT_DURATION = time.Duration(5) * time.Second

	// wait this much time before retrying
	RETRY_BACKOFF_TIME = time.Duration(10) * time.Second

	// maximum amount of retries
	MAX_RETRIES = 3
)

var (
	// maximum ffmpeg processes
	MAXPROCS = runtime.NumCPU()
)
