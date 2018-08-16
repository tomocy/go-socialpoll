package main

import (
	"encoding/json"
	"net/http"
)

func encodeBody(w http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}

func decodeBody(r *http.Request, data interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(data)
}
