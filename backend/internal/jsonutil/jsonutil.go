package jsonutil

import (
	"encoding/json"
	"net/http"
)

func WriteHttpJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, message string, status int) {
	WriteHttpJson(w, status, map[string]string{"error": message})
}
