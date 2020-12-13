package main

import (
	"context"
	"os"

	pkg "github.com/L04DB4L4NC3R/spotify-downloader/ytber/pkg"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

func connectYoutube(ctx context.Context, apikey string) (*youtube.Service, error) {
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		return nil, err
	}
	return youtubeService, nil
}

func redisConnect() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	rdc := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	ctx := context.Background()
	_, err := rdc.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"redis_server": addr,
	}).Info("Connected to Redis")
	return rdc, nil
}

func init() {
	err := godotenv.Load("./config/secret.env")
	if err != nil {
		log.Warn("No env file to load")
	}
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	rdc, err := redisConnect()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	ytSvc, err := connectYoutube(ctx, os.Getenv("YOUTUBE_APIKEY"))
	if err != nil {
		log.Fatal(err)
	}
	if err := pkg.Register(ytSvc, rdc); err != nil {
		log.Fatal(err)
	}
}
