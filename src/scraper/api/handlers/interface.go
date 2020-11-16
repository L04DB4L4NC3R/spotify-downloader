package handler

import (
	"net/http"
)

type Handler interface {
	Health() http.Handler
}
