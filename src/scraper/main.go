package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/handlers"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/core"
	pb "github.com/L04DB4L4NC3R/spotify-downloader/scraper/proto"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rapito/go-spotify/spotify"
	log "github.com/sirupsen/logrus"
)

func redisConnect() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	rdc := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASS"),
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

func spotifyApiConnect() (*spotify.Spotify, error) {
	client := spotify.New(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))
	_, err := client.Authorize()
	if err != nil {
		return nil, err[0]
	}
	log.Info("Connected to Spotify")
	return &client, nil
}

// Global channel pool is being run as a goroutine to listen for events throughout the application
// Additional channels can be added for seperation of concerns when it comes to type of events
func globalChannelPool(cerr chan core.AsyncErrors) {
	select {
	case errobj := <-cerr:
		log.WithFields(log.Fields{
			"error": errobj.Err(),
			"msg":   errobj.Msg(),
			"src":   errobj.Src(),
			"data":  errobj.Data(),
		}).Error("Some error was caught by the async error handler")
	}
}

func init() {
	err := godotenv.Load("./config/scraper.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	if os.Getenv("ENV") == "DEV" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	// create redis repo and redis client
	rdc, err := redisConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer rdc.Close()
	// create redis error handling channel
	cerr := make(chan core.AsyncErrors)
	redisRepo := core.NewRedisRepo(rdc, cerr)
	// create core service using redis repo
	spotifyClient, err := spotifyApiConnect()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// create a gRPC client for ytber
	conn, feedMetaClient, err := pb.Register()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	defer conn.Close()
	coreSvc := core.NewService(redisRepo, spotifyClient, pb.NewService(feedMetaClient))

	// create a router and register handlers
	r := mux.NewRouter()
	handler.RegisterHandler(r, coreSvc)

	// make HTTP server using mux
	addr := "127.0.0.1:3000"
	var rwTimeout time.Duration = 15
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: rwTimeout * time.Second,
		ReadTimeout:  rwTimeout * time.Second,
	}

	log.WithFields(log.Fields{
		"web_server":    addr,
		"write_timeout": rwTimeout,
		"read_timeout":  rwTimeout,
	}).Info("Listening....")

	// start global channel pool
	go globalChannelPool(cerr)
	// graceful shutdown
	cleanup(cerr)

	// start the HTTP server
	log.Fatal(srv.ListenAndServe())
}

func cleanup(cerr chan core.AsyncErrors) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			log.Infoln("Graceful Shutdown Initiated")
			close(cerr)
			log.Infoln("Closed Global Async Error Channel")
			os.Exit(0)
		}
	}()
}
