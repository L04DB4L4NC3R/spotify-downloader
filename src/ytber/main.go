package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	pkg "github.com/L04DB4L4NC3R/spotify-downloader/ytber/pkg"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func globalChannelPool(cerr chan pkg.AsyncErrors) {
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

	cerr := make(chan pkg.AsyncErrors)
	redisRepo := pkg.NewRedisRepo(rdc, cerr)

	go globalChannelPool(cerr)
	cleanup(cerr)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	if err := pkg.Register(redisRepo, cerr); err != nil {
		log.Fatal(err)
	}
}

func cleanup(cerr chan pkg.AsyncErrors) {
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
