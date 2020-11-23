package main

import (
	"os"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {
	err := godotenv.Load("./config/scraper.env")
	if err != nil {
		log.Fatal("Error loading .env file")
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
	if err := pb.Register(); err != nil {
		log.Fatal(err)
	}
}
