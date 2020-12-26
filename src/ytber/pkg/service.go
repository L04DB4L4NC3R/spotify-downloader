package pkg

import (
	"bytes"
	context "context"
	"fmt"
	"html"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
)

type service struct {
	redis Repository
	cerr  chan AsyncErrors
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	offloadToYoutubeDL(ctx context.Context, format string, query string, songmeta *pb.SongMetaRequest, wg *sync.WaitGroup) (postprocessingcmd string)
	offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest, fetchChan chan string) (postprocessingcmds []string)
	applyThumbnailsSerially(ctx context.Context, thumbscmds []string)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": html.EscapeString(req.Title),
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.Title, html.EscapeString(req.ArtistName))

	// no need for this but maintaining it because wg is useful in the case of playlists
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		thumbscmd := s.offloadToYoutubeDL(ctx, "mp3", query, req, wg)
		s.applyThumbnailsSerially(ctx, []string{thumbscmd})
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
	}).Info("Received Playlist Download Request")

	go func(count int) {
		PLAYLIST_BATCH_SIZE, _ := strconv.Atoi(os.Getenv("PLAYLIST_BATCH_SIZE"))
		var (
			batchCount   int     = 1
			totalBatches float64 = math.Ceil(float64(count) / float64(PLAYLIST_BATCH_SIZE))
			cmds         []string
		)

		fetchChan := make(chan string, PLAYLIST_BATCH_SIZE)
		defer close(fetchChan)

		dispatchTime := time.Now()

		for i := 0; i < count; i += PLAYLIST_BATCH_SIZE {
			var offset int = i + PLAYLIST_BATCH_SIZE
			if offset >= count {
				offset = count - 1
			}
			log.WithFields(log.Fields{
				"total_songs_count": count,
				"batch_number":      batchCount,
				"total_batches":     totalBatches,
			}).Info("Playlist Batch Execution")
			results := s.offloadBatchToYoutubeDL(ctx, req.Songs[i:offset], fetchChan)
			cmds = append(cmds, results...)
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

	}(count)

	res := &pb.PlaylistMetaResponse{
		Success: true,
	}
	return res, nil
}

func (s *service) offloadToYoutubeDL(ctx context.Context,
	format string,
	query string,
	songmeta *pb.SongMetaRequest,
	wg *sync.WaitGroup) (postprocessingcmd string) {

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)

	artistName := html.EscapeString(songmeta.ArtistName)
	songTitle := html.EscapeString(songmeta.Title)
	albumName := html.EscapeString(songmeta.AlbumName)

	metacommand := fmt.Sprintf(YT_DOWNLOAD_METADATA_ARGS, artistName, songTitle, songmeta.Date, songmeta.Url, string(songmeta.Track))

	downloadcommand := command + metacommand + YT_DOWNLOAD_PATH_CMD
	// TODO: use CommandContext
	cmd := exec.Command("sh", "-c", downloadcommand)

	var out bytes.Buffer
	var serr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &serr
	if err := cmd.Start(); err != nil {
		s.cerr <- NewRepoError("Error Queuing Job", err, SRC_YTDL, downloadcommand)
		go s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_QUEUED)

	if err := cmd.Wait(); err != nil {
		fmt.Println(serr.String())
		s.cerr <- NewRepoError("Error Executing Job", err, SRC_YTDL, downloadcommand)
		go s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_COMPLETE)
	log.WithFields(log.Fields{
		"song": query,
	}).Info("Download Completed")

	// apply thumbnail

	logs := strings.Split(out.String(), "\n")
	// TODO: prevent panic
	dwpath := strings.Split(logs[len(logs)-3], "[ffmpeg] Adding metadata to '")[1]
	dwpath = dwpath[:len(dwpath)-1]

	// thumbnail command
	postprocessingcmd = fmt.Sprintf(FFMPEG_THUMBNAIL_CMD, songmeta.Thumbnail, dwpath, "music", songTitle, artistName, albumName, dwpath)

	wg.Done()
	return postprocessingcmd
}

func (s *service) offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest, fetchChan chan string) (postprocessingcmds []string) {

	// to see when all the songs of the current batch are downloaded
	batchSize := len(slice)
	songWg := &sync.WaitGroup{}
	songWg.Add(batchSize)

	for _, v := range slice {
		query := fmt.Sprintf("%s - %s", v.Title, v.ArtistName)
		go func(v *pb.SongMetaRequest) {
			fetchChan <- s.offloadToYoutubeDL(ctx, "mp3", query, v, songWg)
		}(v)
	}

	// to maintain atomicity from other batches so that no batch can start executing once this is done
	// wait till all the songs in the current batch are downloaded, then mark the current batch as done
	songWg.Wait()
	for i := 0; i < batchSize; i++ {
		postprocessingcmds = append(postprocessingcmds, <-fetchChan)
	}
	return postprocessingcmds
}

// ffmpeg encoding with thumbnail is best done serially
// after all the songs are downloaded
func (s *service) applyThumbnailsSerially(ctx context.Context, thumbscmds []string) {

	log.Infof("Queuing %d thumbnail jobs", len(thumbscmds))
	command := strings.Join(thumbscmds, ";")
	// TODO: use CommandContext
	cmd := exec.Command("sh", "-c", command)

	if err := cmd.Start(); err != nil {
		s.cerr <- NewRepoError("Error Queuing Thumbnail Job", err, SRC_YTDL, thumbscmds)
		return
	}
	if err := cmd.Wait(); err != nil {
		s.cerr <- NewRepoError("Error Executing Thumbnail Job", err, SRC_YTDL, thumbscmds)
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
