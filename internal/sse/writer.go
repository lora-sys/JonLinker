package sse

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func WriteEvent(w io.Writer, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("data: " + string(b) + "\n\n"))
	return err
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("WriteJSON encode: %v", err)
	}
}
