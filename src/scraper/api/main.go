package main

import (
	"net/http"
	"os"
	"time"

	handler "github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/handlers"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/api/middleware"
	"github.com/L04DB4L4NC3R/spotify-downloader/scraper/core"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func init() {
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
	// create redis repo
	redisRepo := core.NewRedisRepo()
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
	log.Fatal(srv.ListenAndServe())
}
