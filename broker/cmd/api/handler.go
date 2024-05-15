package main

import (
	"broker/data"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChanStruct struct {
	JsReq JsonRequest
	Error error
}

type JsonRequest struct {
	Hash string `json:"hash" bson:"hash"`
}
type JsonResponse struct {
	Error bool   `json:"error"`
	Text  string `json:"text"`
}

//go:embed templates/*
var templateFS embed.FS

func (app *Config) HandleGetMainPage(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	templ.ExecuteTemplate(w, "index", nil)
}

func (app *Config) GetHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	mes, err := data.GetByKey(hash)
	if err == nil {
		log.Println("found via redis")
		var js JsonResponse
		js.Error = false
		js.Text = mes.Text
		templ := template.Must(template.ParseFS(templateFS, "templates/textblock.html.gohtml"))
		templ.ExecuteTemplate(w, "index", js)
		return
	}
	//----------------------via mongo--------------------
	mes, err = data.FindMesByHash(hash)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		templ.ExecuteTemplate(w, "notFound", nil)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error!"))
		return
	}
	err = mes.SetMes()
	if err != nil {
		log.Println("Error during update cache")
	}
	log.Println("found via mongodb")
	var jsresp JsonResponse
	jsresp.Error = false
	jsresp.Text = mes.Text
	templ := template.Must(template.ParseFS(templateFS, "templates/textblock.html.gohtml"))
	templ.ExecuteTemplate(w, "index", jsresp)
}

func (app *Config) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	var mes data.Message
	mes.Text = r.FormValue("text")
	ttl := r.FormValue("htl")
	if mes.Text == "" || ttl == "0" {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	log.Println(mes.Text)
	log.Println(ttl)
	htl, err := strconv.Atoi(ttl)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	mes.HTL = int64(htl)
	resp, err := MakeRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	mes.Hash = resp.Hash
	if err = mes.InsertMes(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	w.WriteHeader(http.StatusOK)
	mes.Hash = fmt.Sprintf("http://localhost:8000/mess%s", mes.Hash)
	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	templ.ExecuteTemplate(w, "link", mes)

}

func GetHashFromHasher() (*JsonRequest, error) {
	req, err := http.NewRequest("POST", "http://hasher/hash", bytes.NewBuffer(nil))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	var resp JsonRequest
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &resp, nil
}

func MakeRequest(r *http.Request) (*JsonRequest, error) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	respch := make(chan ChanStruct)
	go func() {
		resp, err := GetHashFromHasher()
		respch <- ChanStruct{
			JsReq: *resp,
			Error: err,
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("request from hasher took to much time")
		case chresp := <-respch:
			return &chresp.JsReq, chresp.Error
		}
	}
}
