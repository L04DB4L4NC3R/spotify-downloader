package pkg

import (
	context "context"
	"fmt"
	"net"
	"os"
	"os/exec"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	STATUS_META_FED     = "FED"
	STATUS_DWN_QUEUED   = "QUEUED"
	STATUS_DWN_FAILED   = "FAILED"
	STATUS_DWN_COMPLETE = "COMPLETE"

	YT_BASE_URL = "https://youtube.com/watch?v="

	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""
)

type service struct {
	redis Repository
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	offloadToYoutubeDL(ctx context.Context, format string, query string, songId string)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": req.Title,
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.Title, req.ArtistName)

	go s.offloadToYoutubeDL(ctx, "mp3", query, req.SongId)
	res := &pb.SongMetaResponse{
		Success: true,
	}
	return res, nil
}

func (s *service) PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error) {
	// TODO: Do stuff
	log.WithFields(log.Fields{
		"count": len(req.Songs),
	}).Info("Received Playlist Download Request")
	res := &pb.PlaylistMetaResponse{
		Success: true,
		ErrMsgs: []string{},
		YtUrls:  []string{},
	}
	return res, nil
}

func (s *service) offloadToYoutubeDL(ctx context.Context,
	format string,
	query string,
	songId string) {

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)
	cmd := exec.Command("bash", "-c", command)

	if err := cmd.Start(); err != nil {
		go s.redis.UpdateStatus("song", songId, STATUS_DWN_FAILED)
		return
	}

	go s.redis.UpdateStatus("song", songId, STATUS_DWN_QUEUED)

	if err := cmd.Wait(); err != nil {
		go s.redis.UpdateStatus("song", songId, STATUS_DWN_FAILED)
		return
	}

	go s.redis.UpdateStatus("song", songId, STATUS_DWN_COMPLETE)
	log.WithFields(log.Fields{
		"song": query,
	}).Info("Download Completed")
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
