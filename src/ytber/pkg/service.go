package pkg

import (
	context "context"
	"net"
	"os"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type service struct {
	redis Repository
	cerr  chan AsyncErrors
}

type Service interface {
	SongDownload(
		ctx context.Context,
		req *pb.SongMetaRequest,
	) (*pb.SongMetaResponse, error)

	PlaylistDownload(
		ctx context.Context,
		req *pb.PlaylistMetaRequest,
	) (*pb.PlaylistMetaResponse, error)

	ApplyThumbnails(
		ctx context.Context,
		thumbscmds []string,
		resources []*pb.SongMetaRequest,
	)
}

func Register(rdc Repository, cerr chan AsyncErrors) error {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	if addr == "" {
		addr = "localhost:4040"
	}
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
