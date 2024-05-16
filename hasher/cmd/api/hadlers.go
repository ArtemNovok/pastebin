package main

import (
	"encoding/json"
	"hasher/data"
	"log"
	"net/http"
)

type JsonRequest struct {
	Hash string `json:"hash" bson:"hash"`
}

func (app *Config) GetHash(w http.ResponseWriter, r *http.Request) {
	hash, err := data.GetDelKey()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get hash from redis"))
		return
	}
	var resp JsonRequest
	resp.Hash = hash
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal response"))
		return
	}
	w.WriteHeader(http.StatusOK)
	size, err := data.GetDBSize()
	if err != nil {
		log.Panic(err)
		return
	}
	if size < 20 {
		go data.GenerateHashes()
		// if err != nil {
		// 	log.Panic(err)
		// 	return
		// }
	}

}
