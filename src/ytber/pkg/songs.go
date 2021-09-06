package pkg

import (
	"context"
	"fmt"
	"html"
	"os/exec"
	"strings"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
)

const (
	htmlSingleQuoteString = "&#39;"
	htmlDoubleQuoteString = "&#34;"
	htmlAmpersand         = "&amp;"
)

func (s *service) SongDownload(ctx context.Context, req *pb.SongMetaRequest) (*pb.SongMetaResponse, error) {
	log.WithFields(log.Fields{
		"url":   req.Url,
		"title": html.EscapeString(req.Title),
	}).Info("Received SongDownload Request")

	query := fmt.Sprintf("%s - %s", req.Title, html.EscapeString(req.ArtistName))

	go func() {
		thumbscmd := s.offloadToYoutubeDL(ctx, query, req)
		s.ApplyThumbnails(ctx, []string{thumbscmd.postprocessingcmd}, []*pb.SongMetaRequest{req})
		s.redis.UpdateStatus(RESOURCE_SONG, req.SongId, STATUS_FINISHED)
	}()

	res := &pb.SongMetaResponse{
		Success: true,
	}
	return res, nil
}

func (s *service) offloadToYoutubeDL(
	nctx context.Context,
	query string,
	songmeta *pb.SongMetaRequest,
) asyncReturn {
	ctx, cancel := context.WithTimeout(context.Background(), SONG_DOWNLOAD_TIMEOUT)
	defer cancel()

	command := fmt.Sprintf(YT_DOWNLOAD_CMD, songmeta.Format, query)

	artistName := html.EscapeString(songmeta.ArtistName)
	songTitle := html.EscapeString(songmeta.Title)
	albumName := html.EscapeString(songmeta.AlbumName)

	errcmd := ""

	metacommand := fmt.Sprintf(YT_DOWNLOAD_METADATA_ARGS, artistName, songTitle, songmeta.Date, songmeta.Url, string(songmeta.Track))
	//metacommand = strings.ReplaceAll(metacommand, htmlSingleQuoteString, "\\'")

	downloadcommand := command + metacommand + YT_DOWNLOAD_PATH_CMD
	//downloadcommand = strings.ReplaceAll(downloadcommand, htmlSingleQuoteString, "\\'")
	s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_QUEUED)

	cmd := exec.CommandContext(ctx, "sh", "-c", downloadcommand)
	out, err := cmd.Output()
	if err != nil {
		//s.cerr <- NewRepoError("Error Executing Job", err, SRC_YTDL, downloadcommand)
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("error executing download")
		s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_FAILED)
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}

	s.redis.UpdateStatus(RESOURCE_SONG, songmeta.SongId, STATUS_DWN_COMPLETE)
	log.WithFields(log.Fields{
		"song": query,
	}).Info("download completed")

	// apply thumbnail
	logs := strings.Split(string(out), "\n")
	// Prevent panic in the case of faulty logs
	if len(logs) < 4 {
		log.WithFields(log.Fields{
			"song": songTitle,
		}).Info("skipping thumbnail application")
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}
	dwpathi := strings.Split(logs[len(logs)-3], "[ffmpeg] Adding metadata to '")
	if len(dwpathi) < 2 {
		log.WithFields(log.Fields{
			"song": songTitle,
		}).Info("skipping thumbnail application")
		return asyncReturn{
			query:             query,
			meta:              songmeta,
			postprocessingcmd: errcmd,
		}
	}

	dwpath := dwpathi[1]
	dwpath = dwpath[:len(dwpath)-1]

	// thumbnail command
	postprocessingcmd := fmt.Sprintf(FFMPEG_THUMBNAIL_CMD, songmeta.Thumbnail, dwpath, "music", songTitle, artistName, albumName, songmeta.Format, dwpath)
	postprocessingcmd = strings.NewReplacer(
		htmlSingleQuoteString, "'",
		htmlDoubleQuoteString, "\"",
		htmlAmpersand, "&",
	).Replace(postprocessingcmd)
	return asyncReturn{
		query:             query,
		meta:              songmeta,
		postprocessingcmd: postprocessingcmd,
	}
}
