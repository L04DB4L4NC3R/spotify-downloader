package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"url":        r.URL.String(),
			"headers":    r.Header.Clone().Values,
			"host":       r.Host,
			"User_agent": r.UserAgent(),
		}).Info("Request Logger")
		h.ServeHTTP(w, r)
	})
}

func RegisterMiddlewares(r *mux.Router) {
	r.Use(Logger)
}
