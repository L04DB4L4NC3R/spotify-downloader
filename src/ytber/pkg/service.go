package pkg

import (
	"bytes"
	context "context"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	STATUS_META_FED     = "FED"
	STATUS_DWN_QUEUED   = "QUEUED"
	STATUS_DWN_FAILED   = "FAILED"
	STATUS_DWN_COMPLETE = "COMPLETED"

	YT_BASE_URL = "https://youtube.com/watch?v="

	// TODO: for docker bind mount
	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""

	YT_DOWNLOAD_METADATA_ARGS = " --add-metadata --postprocessor-args '-metadata artist=\"%s\" -metadata title=\"%s\" -metadata date=\"%s\" -metadata purl=\"%s\" -metadata track=\"%s\"'"

	YT_DOWNLOAD_PATH_CMD = " -o \"music/%(title)s.%(ext)s\""

	// image url
	// song path
	// download path
	// title
	// song path
	FFMPEG_THUMBNAIL_CMD = "ffmpeg -i %s -i \"%s\" -map_metadata 1 -map 1 -map 0 \"%s/%s.mp3\" && rm \"%s\""
)

type service struct {
	redis Repository
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	offloadToYoutubeDL(ctx context.Context, format string, query string, songmeta *pb.SongMetaRequest, wg *sync.WaitGroup)
	offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": req.Title,
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.Title, req.ArtistName)

	// no need for this but maintaining it because wg is useful in the case of playlists
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.offloadToYoutubeDL(ctx, "mp3", query, req, wg)
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
		)
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
			s.offloadBatchToYoutubeDL(ctx, req.Songs[i:offset])
			batchCount++
		}
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
	wg *sync.WaitGroup) {

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)

	metacommand := fmt.Sprintf(YT_DOWNLOAD_METADATA_ARGS, songmeta.ArtistName, songmeta.Title, songmeta.Date, songmeta.Url, string(songmeta.Track))

	downloadcommand := command + metacommand + YT_DOWNLOAD_PATH_CMD
	cmd := exec.Command("sh", "-c", downloadcommand)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_QUEUED)

	if err := cmd.Wait(); err != nil {
		go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus("song", songmeta.SongId, STATUS_DWN_COMPLETE)
	log.WithFields(log.Fields{
		"song": query,
	}).Info("Download Completed")

	// TODO: apply thumbnail (make a status for that also in redis)
	// apply thumbnail
	logs := strings.Split(out.String(), "\n")
	dwpath := strings.Split(logs[len(logs)-3], "[ffmpeg] Adding metadata to '")[1]
	dwpath = dwpath[:len(dwpath)-1]

	thumbscommand := fmt.Sprintf(FFMPEG_THUMBNAIL_CMD, songmeta.Thumbnail, dwpath, "music", songmeta.Title, dwpath)
	fmt.Println(thumbscommand)
	cmd = exec.Command("sh", "-c", thumbscommand)

	if err := cmd.Start(); err != nil {
		wg.Done()
		return
	}
	if err := cmd.Wait(); err != nil {
		wg.Done()
		return
	}
	log.WithFields(log.Fields{
		"song": songmeta.Title,
	}).Info("Thumbnail Applied")
	wg.Done()
}

func (s *service) offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest) {

	// to see when all the songs of the current batch are downloaded
	songWg := &sync.WaitGroup{}
	songWg.Add(len(slice))

	for _, v := range slice {
		query := fmt.Sprintf("%s - %s", v.Title, v.ArtistName)
		go s.offloadToYoutubeDL(ctx, "mp3", query, v, songWg)
	}

	// to maintain atomicity from other batches so that no batch can start executing once this is done
	// wait till all the songs in the current batch are downloaded, then mark the current batch as done
	songWg.Wait()
}

func Register(rdc Repository) error {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterFeedMetaServer(srv, &service{rdc})
	log.WithFields(log.Fields{
		"grpc_server": addr,
	}).Info("Started gRPC server")
	if err := srv.Serve(listener); err != nil {
		return err
	}
	return nil
}
