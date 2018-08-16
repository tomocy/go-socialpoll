package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, data)
	}
}

func respondErr(w http.ResponseWriter, status int, data ...interface{}) {
	respond(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(data...),
		},
	})
}

func encodeBody(w http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}

func decodeBody(r *http.Request, data interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(data)
}
