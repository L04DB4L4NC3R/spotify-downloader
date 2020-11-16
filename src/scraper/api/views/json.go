package views

import (
	"encoding/json"
	"net/http"
)

func Fill(w http.ResponseWriter, message string, data interface{}, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
}
