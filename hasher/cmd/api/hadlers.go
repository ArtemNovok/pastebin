package main

import (
	"encoding/json"
	"hasher/data"
	"net/http"
	"strconv"

	"github.com/shomali11/util/xhashes"
)

type JsonRequest struct {
	Hash string `json:"hash" bson:"hash"`
}

func (app *Config) GetHash(w http.ResponseWriter, r *http.Request) {
	var req JsonRequest
	num, err := data.GetNum()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to get num from db"))
		return
	}
	hs := strconv.FormatUint(uint64(xhashes.FNV32(string(num))), 10)
	req.Hash = hs
	js, err := json.Marshal(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to hash"))
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write(js)
}
