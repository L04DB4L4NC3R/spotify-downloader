package pkg

import (
	context "context"
	"fmt"
	"html"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	STATUS_META_FED     = "FED"
	STATUS_DWN_QUEUED   = "QUEUED"
	STATUS_DWN_FAILED   = "FAILED"
	STATUS_DWN_COMPLETE = "COMPLETED" // song downloaded (before thumbnail application)
	STATUS_FINISHED     = "FINISHED"  // thumbnail applied

	YT_BASE_URL = "https://youtube.com/watch?v="

	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""

	YT_DOWNLOAD_METADATA_ARGS = " --add-metadata --postprocessor-args $'-metadata artist=\"%s\" -metadata title=\"%s\" -metadata date=\"%s\" -metadata purl=\"%s\" -metadata track=\"%s\"'"

	YT_DOWNLOAD_PATH_CMD = " -o \"music/%(title)s.%(ext)s\""

	// image url
	// song path
	// download path
	// title
	// song path
	FFMPEG_THUMBNAIL_CMD = "ffmpeg -y -i %s -i \"%s\" -map_metadata 1 -map 1 -map 0 \"%s/%s -(%s)-(%s).mp3\" && rm \"%s\""

	RESOURCE_PLAYLIST = "playlists"
	RESOURCE_SONG     = "tracks"
	RESOURCE_ALBUM    = "albums"

	SONG_DOWNLOAD_TIMEOUT         = time.Duration(2) * time.Minute
	BATCH_DOWNLOAD_TIMEOUT        = time.Duration(5) * time.Minute
	THUMBNAIL_APPLICATION_TIMEOUT = time.Duration(20) * time.Minute
	SONG_DOWNLOAD_WAIT_DURATION   = time.Duration(10) * time.Second
	RETRY_BACKOFF_TIME            = time.Duration(10) * time.Second

	MAX_RETRIES = 3
)

type service struct {
	redis Repository
	cerr  chan AsyncErrors
}

type asyncReturn struct {
	meta              *pb.SongMetaRequest
	query             string
	postprocessingcmd string
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	offloadToYoutubeDL(ctx context.Context, format string, query string, songmeta *pb.SongMetaRequest) (postprocessingcmd asyncReturn)
	offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest, fetchChan chan asyncReturn) (postprocessingcmds []asyncReturn, retries []*pb.SongMetaRequest)
	applyThumbnailsSerially(ctx context.Context, thumbscmds []string)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": html.EscapeString(req.Title),
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.Title, html.EscapeString(req.ArtistName))

	go func() {
		thumbscmd := s.offloadToYoutubeDL(ctx, "mp3", query, req)
		s.applyThumbnailsSerially(ctx, []string{thumbscmd.postprocessingcmd})
		s.redis.UpdateStatus(RESOURCE_SONG, req.SongId, STATUS_FINISHED)
	}()

	res := &pb.SongMetaResponse{
		Success: true,
	}
	return res, nil
}

func (s *service) PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error) {
	var count int = len(req.Songs)
	log.WithFields(log.Fields{
		"count": count,
	}).Info("received playlist download request")

	go func(count int) {
		PLAYLIST_BATCH_SIZE, _ := strconv.Atoi(os.Getenv("PLAYLIST_BATCH_SIZE"))
		var (
			batchCount   int     = 1
			retryCount   int     = 0
			totalBatches float64 = math.Ceil(float64(count) / float64(PLAYLIST_BATCH_SIZE))
			cmds         []string
			retries      []*pb.SongMetaRequest
		)

		fetchChan := make(chan asyncReturn, PLAYLIST_BATCH_SIZE)
		defer close(fetchChan)

		dispatchTime := time.Now()

		for i := 0; i < count; i += PLAYLIST_BATCH_SIZE {
			var offset int = i + PLAYLIST_BATCH_SIZE
			if offset > count {
				offset = count
			}
			log.WithFields(log.Fields{
				"total_songs_count": count,
				"batch_number":      batchCount,
				"total_batches":     totalBatches,
				"batch_index":       i,
				"batch_offset":      offset,
			}).Info("playlist batch execution")
			results, retry := s.offloadBatchToYoutubeDL(ctx, req.Songs[i:offset], fetchChan)
			cmds = append(cmds, results...)
			retries = append(retries, retry...)
			batchCount++
		}

		log.WithFields(log.Fields{
			"total_songs_count": count,
			"total_batches":     totalBatches,
			"batch_size":        PLAYLIST_BATCH_SIZE,
			"time_taken":        time.Since(dispatchTime).Minutes(),
		}).Info("Download Successful")
		s.redis.UpdateStatus(req.Type, req.ResourceId, STATUS_DWN_COMPLETE)
		s.applyThumbnailsSerially(ctx, cmds)
		s.redis.UpdateStatus(req.Type, req.ResourceId, STATUS_FINISHED)

		// using a closure for retry
		func() {
			numRetries := len(retries)
			if numRetries != 0 && retryCount < MAX_RETRIES {
				retryCount++
				log.WithFields(log.Fields{
					"retry_backoff_time_seconds": RETRY_BACKOFF_TIME.Seconds(),
					"number_of_songs":            numRetries,
					"current_retry":              retryCount,
					"max_retry_count":            MAX_RETRIES,
				})
				time.Sleep(RETRY_BACKOFF_TIME)
				_, _ = s.PlaylistDownload(
					ctx,
					&pb.PlaylistMetaRequest{
						Songs:      retries,
						ResourceId: req.ResourceId,
						Type:       req.Type,
					},
				)
			}
		}()
	}(count)

	res := &pb.PlaylistMetaResponse{
		Success: true,
	}

	return res, nil
}

func (s *service) offloadToYoutubeDL(
	ctx context.Context,
	format string,
	query string,
	songmeta *pb.SongMetaRequest,
) asyncReturn {

	ctx, _ = context.WithTimeout(context.Background(), SONG_DOWNLOAD_TIMEOUT)

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)

	artistName := html.EscapeString(songmeta.ArtistName)
	songTitle := html.EscapeString(songmeta.Title)
	albumName := html.EscapeString(songmeta.AlbumName)

	errcmd := ""

	metacommand := fmt.Sprintf(YT_DOWNLOAD_METADATA_ARGS, artistName, songTitle, songmeta.Date, songmeta.Url, string(songmeta.Track))

	downloadcommand := command + metacommand + YT_DOWNLOAD_PATH_CMD
	go s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_QUEUED)

	cmd := exec.CommandContext(ctx, "sh", "-c", downloadcommand)
	out, err := cmd.Output()
	if err != nil {
		//s.cerr <- NewRepoError("Error Executing Job", err, SRC_YTDL, downloadcommand)
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("error executing download")
		go s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_FAILED)
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}

	go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_COMPLETE)
	log.WithFields(log.Fields{
		"song": query,
	}).Info("download completed")

	// apply thumbnail
	logs := strings.Split(string(out), "\n")
	// Prevent panic in the case of faulty logs
	if len(logs) < 4 {
		log.WithFields(log.Fields{
			"song": songTitle,
		}).Info("skipping thumbnail application")
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}
	dwpathi := strings.Split(logs[len(logs)-3], "[ffmpeg] Adding metadata to '")
	if len(dwpathi) < 2 {
		log.WithFields(log.Fields{
			"song": songTitle,
		}).Info("skipping thumbnail application")
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}

	dwpath := dwpathi[1]
	dwpath = dwpath[:len(dwpath)-1]

	// thumbnail command
	postprocessingcmd := fmt.Sprintf(FFMPEG_THUMBNAIL_CMD, songmeta.Thumbnail, dwpath, "music", songTitle, artistName, albumName, dwpath)
	return asyncReturn{
		query:             query,
		meta:              songmeta,
		postprocessingcmd: postprocessingcmd,
	}
}

func (s *service) offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest, fetchChan chan asyncReturn) (postprocessingcmds []string, retry []*pb.SongMetaRequest) {

	ctx, cancel := context.WithTimeout(context.Background(), BATCH_DOWNLOAD_TIMEOUT)
	defer cancel()

	// to see when all the songs of the current batch are downloaded
	var (
		batchSize = len(slice)
	)

	for _, v := range slice {
		go func(v *pb.SongMetaRequest) {
			query := fmt.Sprintf("%s - %s", v.Title, v.ArtistName)
			fetchChan <- s.offloadToYoutubeDL(ctx, "mp3", query, v)
		}(v)
	}

	count := 0
	for {
		select {
		case <-ctx.Done():
			return postprocessingcmds, retry
		case result := <-fetchChan:
			if count == batchSize-1 {
				return postprocessingcmds, retry
			}
			if result.postprocessingcmd != "" {
				postprocessingcmds = append(postprocessingcmds, result.postprocessingcmd)
			} else {
				retry = append(retry, result.meta)
			}
			count++
		default:
			time.Sleep(SONG_DOWNLOAD_WAIT_DURATION)
		}
	}
}

// ffmpeg encoding with thumbnail is best done serially
// after all the songs are downloaded
func (s *service) applyThumbnailsSerially(ctx context.Context, thumbscmds []string) {

	log.Infof("Queuing %d thumbnail jobs", len(thumbscmds))
	command := strings.Join(thumbscmds, ";")

	ctx, _ = context.WithTimeout(context.Background(), THUMBNAIL_APPLICATION_TIMEOUT)
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("error executing thumbnail application")
		return
	}

	log.Info("Thumbnails Applied")
}

func Register(rdc Repository, cerr chan AsyncErrors) error {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterFeedMetaServer(srv, &service{rdc, cerr})
	log.WithFields(log.Fields{
		"grpc_server": addr,
	}).Info("Started gRPC server")
	if err := srv.Serve(listener); err != nil {
		return err
	}
	return nil
}
