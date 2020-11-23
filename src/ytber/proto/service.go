package pb

import (
	context "context"

	log "github.com/sirupsen/logrus"
)

type service struct {
}

type Service interface {
	SongDownload(ctx context.Context, req *SongMetaRequest) (*SongMetaResponse, error)
	PlaylistDownload(ctx context.Context, req *PlaylistMetaRequest) (*PlaylistMetaResponse, error)
}

func (s *service) SongDownload(ctx context.Context, req *SongMetaRequest) (*SongMetaResponse, error) {
	// TODO: Do stuff
	log.WithFields(log.Fields{
		"url": req.Url,
	}).Info("Received SongDownload Request")
	res := &SongMetaResponse{
		Success: true,
		ErrMsg:  "",
		YtUrl:   "",
	}
	return res, nil
}

func (s *service) PlaylistDownload(ctx context.Context, req *PlaylistMetaRequest) (*PlaylistMetaResponse, error) {
	// TODO: Do stuff
	log.WithFields(log.Fields{
		"count": len(req.Songs),
	}).Info("Received Playlist Download Request")
	res := &PlaylistMetaResponse{
		Success: true,
		ErrMsgs: []string{},
		YtUrls:  []string{},
	}
	return res, nil
}
