package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/handlers"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/middleware"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/core"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func cleanup(rdc *redis.Client, cerr chan core.AsyncErrors) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(rdc *redis.Client, cerr chan core.AsyncErrors) {
		select {
		case <-c:
			log.Infoln("Graceful Shutdown Initiated")
			if err := rdc.Close(); err != nil {
				log.Error(err)
			}
			log.Infoln("Closed Redis Connection")
			close(cerr)
			log.Infoln("Closed Global Async Error Channel")
			os.Exit(0)
		}
	}(rdc, cerr)
}

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
	log.Info("Connected to Redis @ " + addr)
	return rdc, nil
}

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
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

}

func registerHandlers(r *mux.Router, svc core.Service) {
	coreHandler := handler.NewHandler(r, svc)
	middleware.RegisterMiddlewares(r)
	r.Handle("/ping", coreHandler.Health())
}

func main() {
	// create redis repo and redis client
	rdc, err := redisConnect()
	if err != nil {
		log.Fatal(err)
	}
	// create redis error handling channel
	cerr := make(chan core.AsyncErrors)
	redisRepo := core.NewRedisRepo(rdc, cerr)
	// create core service using redis repo
	coreSvc := core.NewService(redisRepo)

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
		"addr":          addr,
		"write_timeout": rwTimeout,
		"read_timeout":  rwTimeout,
	}).Info("Listening....")

	// start global channel pool
	go globalChannelPool(cerr)
	// graceful shutdown
	cleanup(rdc, cerr)

	// start the HTTP server
	log.Fatal(srv.ListenAndServe())
}
