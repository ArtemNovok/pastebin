package main

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, staus int, message string) error {
	w.WriteHeader(staus)
	data := JsonResponse{
		Error:   true,
		Message: message,
	}
	err := json.NewEncoder(w).Encode(&data)
	if err != nil {
		return err
	}
	return nil
}
