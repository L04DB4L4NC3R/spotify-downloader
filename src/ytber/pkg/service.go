package pkg

import (
	"bytes"
	context "context"
	"fmt"
	"net"
	"os"
	"os/exec"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	YT_BASE_URL     = "https://youtube.com/watch?v="
	YT_DOWNLOAD_CMD = "youtube-dl -x --audio-format %s --prefer-ffmpeg --default-search \"ytsearch\" \"%s\""
)

type service struct {
	redisClient *redis.Client
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	OffloadToYoutubeDL(ctx context.Context, format string, query string)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	// TODO: Do stuff
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": req.Title,
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.ArtistName, req.AlbumName)
	log.Info(query)
	go s.offloadToYoutubeDL(ctx, "mp3", query)
	res := &pb.SongMetaResponse{
		Success: true,
		ErrMsg:  "",
		YtUrl:   "",
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

func (s *service) offloadToYoutubeDL(ctx context.Context, format string, query string) {

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, format, query)
	cmd := exec.Command("bash", "-c", command)

	var out bytes.Buffer
	cmd.Stdout = &out

	var er bytes.Buffer
	cmd.Stderr = &er
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	log.Printf("translated phrase: %q\n", out.String())
	log.Printf("error phrase: %q\n", er.String())
}

func Register(rdc *redis.Client) error {
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
