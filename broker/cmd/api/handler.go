package main

import (
	"broker/data"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type JsonRequest struct {
	Hash string `json:"hash" bson:"hash"`
}
type JsonResponse struct {
	Error bool   `json:"error"`
	Text  string `json:"text"`
}

func (app *Config) HandleGetMainPage(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("./templates/main.html.gohtml"))
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
		templ := template.Must(template.ParseFiles("./templates/textblock.html.gohtml"))
		templ.ExecuteTemplate(w, "index", js)
		return
	}
	//----------------------via mongo--------------------
	mes, err = data.FindMesByHash(hash)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFiles("./templates/main.html.gohtml"))
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
	templ := template.Must(template.ParseFiles("./templates/textblock.html.gohtml"))
	templ.ExecuteTemplate(w, "index", jsresp)
}

func (app *Config) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	var mes data.Message
	mes.Text = r.FormValue("text")
	ttl := r.FormValue("htl")
	log.Println(mes.Text)
	log.Println(ttl)
	htl, err := strconv.Atoi(ttl)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to decode message"))
		return
	}
	mes.HTL = int64(htl)
	log.Println(mes)
	req, err := http.NewRequest("POST", "http://localhost:80/hash", bytes.NewBuffer(nil))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to connect  to hash service"))
		return
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to connect  to hash service"))
		return
	}
	defer res.Body.Close()
	var resp JsonRequest
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to decode response from hasher "))
		return
	}
	mes.Hash = resp.Hash
	if err = mes.InsertMes(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to save message to db"))
		return
	}
	w.WriteHeader(http.StatusOK)
	mes.Hash = fmt.Sprintf("http://localhost:8000/mess%s", mes.Hash)
	templ := template.Must(template.ParseFiles("./templates/main.html.gohtml"))
	templ.ExecuteTemplate(w, "link", mes)

}
