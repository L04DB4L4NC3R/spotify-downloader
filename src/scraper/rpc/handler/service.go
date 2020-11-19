package rpcservice

import (
	pb "github.com/L04DB4L4NC3R/spotify-downloader/scraper/rpc/proto"
)

type service struct {
	metaTransportClient *pb.FeedMetaClient
}

type Service interface {
	SendSongMeta(map[string]interface{}) (ytlink string, err error)
	SendPlaylistMeta([]map[string]interface{}) (ytlinks []string, err []error)
}

func NewService(mtc *pb.FeedMetaClient) Service {
	return &service{
		metaTransportClient: mtc,
	}
}

func (svc *service) SendSongMeta(map[string]interface{}) (ytlink string, err error) {
	panic("not implemented")
}

func (svc *service) SendPlaylistMeta([]map[string]interface{}) (ytlinks []string, err []error) {
	panic("not implemented")
}
