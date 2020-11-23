package pb

import (
	context "context"
	"errors"
)

type service struct {
	feedMetaTransporter FeedMetaClient
}

type Service interface {
	SendSongMeta(*songMetaStruct) (ytlink string, err error)
	SendPlaylistMeta([]songMetaStruct) (ytlinks []string, err []error)
	NewSongMetaTransportStruct(url, songId, thumbnail, genre, date, albumUrl, albumName, artistLink, artistName string, duration, track uint32) *songMetaStruct
}

func NewService(mtc FeedMetaClient) Service {
	return &service{
		feedMetaTransporter: mtc,
	}
}

func (svc *service) NewSongMetaTransportStruct(url, songId, thumbnail, genre, date, albumUrl, albumName, artistLink, artistName string, duration, track uint32) *songMetaStruct {
	return &songMetaStruct{
		Url:        url,
		SongId:     songId,
		Thumbnail:  thumbnail,
		Genre:      genre,
		Date:       date,
		AlbumUrl:   albumUrl,
		AlbumName:  albumName,
		ArtistLink: artistLink,
		ArtistName: artistName,
		Duration:   duration,
		Track:      track,
	}
}
func (svc *service) SendSongMeta(meta *songMetaStruct) (ytlink string, err error) {

	req := &SongMetaRequest{
		Url:        meta.Url,
		SongId:     meta.SongId,
		Thumbnail:  meta.Thumbnail,
		Genre:      meta.Genre,
		Date:       meta.Date,
		AlbumUrl:   meta.AlbumUrl,
		AlbumName:  meta.AlbumName,
		ArtistLink: meta.ArtistLink,
		ArtistName: meta.ArtistName,
		Duration:   meta.Duration,
		Track:      meta.Track,
	}
	res, err := svc.feedMetaTransporter.SongDownload(context.Background(), req)
	if err != nil {
		return "", nil
	}
	if ok := res.Success; !ok {
		return "", errors.New(res.ErrMsg)
	}
	return res.YtUrl, nil
}

func (svc *service) SendPlaylistMeta(metas []songMetaStruct) ([]string, []error) {

	var (
		errs    []error
		results []string
	)
	for _, meta := range metas {
		req := &SongMetaRequest{
			Url:        meta.Url,
			SongId:     meta.SongId,
			Thumbnail:  meta.Thumbnail,
			Genre:      meta.Genre,
			Date:       meta.Date,
			AlbumUrl:   meta.AlbumUrl,
			AlbumName:  meta.AlbumName,
			ArtistLink: meta.ArtistLink,
			ArtistName: meta.ArtistName,
			Duration:   meta.Duration,
			Track:      meta.Track,
		}
		res, err := svc.feedMetaTransporter.SongDownload(context.Background(), req)
		if err != nil {
			return nil, []error{err}
		}
		if ok := res.Success; !ok {
			errs = append(errs, errors.New(res.ErrMsg))
			continue
		}
		results = append(results, res.YtUrl)
	}
	return results, errs
}
