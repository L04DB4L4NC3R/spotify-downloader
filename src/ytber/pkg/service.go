package pkg

import (
	context "context"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
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

	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""
)

type service struct {
	redis Repository
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	offloadToYoutubeDL(ctx context.Context, format string, query string, songId string, wg *sync.WaitGroup)
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
	go s.offloadToYoutubeDL(ctx, "mp3", query, req.SongId, wg)
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
	songId string,
	wg *sync.WaitGroup) {

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)
	cmd := exec.Command("sh", "-c", command)

	if err := cmd.Start(); err != nil {
		go s.redis.UpdateStatus("song", songId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus("song", songId, STATUS_DWN_QUEUED)

	if err := cmd.Wait(); err != nil {
		go s.redis.UpdateStatus("song", songId, STATUS_DWN_FAILED)
		wg.Done()
		return
	}

	go s.redis.UpdateStatus("song", songId, STATUS_DWN_COMPLETE)
	wg.Done()
	log.WithFields(log.Fields{
		"song": query,
	}).Info("Download Completed")
}

func (s *service) offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest) {

	// to see when all the songs of the current batch are downloaded
	songWg := &sync.WaitGroup{}
	songWg.Add(len(slice))

	for _, v := range slice {
		query := fmt.Sprintf("%s - %s", v.Title, v.ArtistName)
		go s.offloadToYoutubeDL(ctx, "mp3", query, v.SongId, songWg)
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
