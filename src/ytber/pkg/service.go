package pkg

import (
	context "context"
	"net"
	"os"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/grpc"
)

const (
	YT_BASE_URL = "https://youtube.com/watch?v="
)

type service struct {
	ytSvc *youtube.Service
}

type Service interface {
	SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error)
	OffloadToYoutubeDL(ctx context.Context, videoId string)
}

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	// TODO: Do stuff
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": req.Title,
	}).Info("Received SongDownload Request")

	query := []string{req.ArtistName, req.AlbumName}
	resp, err := s.ytSvc.Search.List(query).MaxResults(1).Do()
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	videoId := resp.Items[0].Id.VideoId
	go s.offloadToYoutubeDL(ctx, videoId)
	res := &pb.SongMetaResponse{
		Success: true,
		ErrMsg:  "",
		YtUrl:   videoId,
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

func (s *service) offloadToYoutubeDL(ctx context.Context, videoId string) {
	url := YT_BASE_URL + videoId
	log.Info(url)
	// TODO: offlaod to youtubedl
}

func Register(ytSvc *youtube.Service) error {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterFeedMetaServer(srv, &service{ytSvc})
	log.WithFields(log.Fields{
		"grpc_server": addr,
	}).Info("Started gRPC server")
	if err := srv.Serve(listener); err != nil {
		return err
	}
	return nil
}
