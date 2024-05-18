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

type ChanHashStruct struct {
	JsReq JsonHashRequest
	Error error
}

type ChanAuthStruct struct {
	JsReq JsonAuthResponse
	Error error
}

type User struct {
	Id       int64  `json:"id,omitempty"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JsonAuthResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type JsonHashRequest struct {
	Hash string `json:"hash" bson:"hash"`
}
type JsonResponse struct {
	Error    bool   `json:"error"`
	Text     string `json:"text"`
	Username string `json:"username"`
	UserId   int64  `json:"userid"`
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
		js.UserId = mes.UserId
		js.Username = mes.UserName
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
		log.Println("Error during cache update ")
	}
	log.Println("found via mongodb")
	var jsresp JsonResponse
	jsresp.Error = false
	jsresp.Text = mes.Text
	jsresp.UserId = mes.UserId
	jsresp.Username = mes.UserName
	templ := template.Must(template.ParseFS(templateFS, "templates/textblock.html.gohtml"))
	templ.ExecuteTemplate(w, "index", jsresp)
}

func (app *Config) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	var mes data.Message
	reqdata := r.Context().Value("userdata").(ReqUserData)
	if reqdata.UserData["Authorized"] == "0" {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	id := reqdata.UserData["Id"]
	username := reqdata.UserData["Username"]
	log.Println(id)
	log.Println(username)
	intId, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	mes.UserId = int64(intId)
	mes.UserName = username
	mes.Text = r.FormValue("text")
	ttl := r.FormValue("htl")
	if mes.Text == "" || ttl == "0" {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	htl, err := strconv.Atoi(ttl)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	mes.HTL = int64(htl)
	resp, err := MakeHashRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
		mes.Error = true
		templ.ExecuteTemplate(w, "link", mes)
		return
	}
	mes.Hash = resp.Hash
	if err = mes.InsertMes(); err != nil {
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

func GetHashFromHasher() (*JsonHashRequest, error) {
	req, err := http.NewRequest("POST", "http://hasher/hash", bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var resp JsonHashRequest
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &resp, nil
}

func MakeHashRequest(r *http.Request) (*JsonHashRequest, error) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	respch := make(chan ChanHashStruct)
	go func() {
		resp, err := GetHashFromHasher()
		respch <- ChanHashStruct{
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

func (app *Config) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	username := r.FormValue("username")

	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	var errResp JsonAuthResponse
	errResp.Error = true
	if email == "" || password == "" || username == "" {
		w.WriteHeader(http.StatusOK)
		errResp.Message = "Invalid credentials!"
		templ.ExecuteTemplate(w, "login", errResp)
		return
	}
	resp, err := MakeSignInRequest(r, email, password, username)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		errResp.Message = "Failed to save user try another time"
		templ.ExecuteTemplate(w, "login", errResp)
		return
	}
	templ.ExecuteTemplate(w, "login", resp)
}

func CreateUser(email, password, username string) (JsonAuthResponse, error) {
	reqdata := User{
		Email:    email,
		Password: password,
		UserName: username,
	}
	payload, err := json.Marshal(reqdata)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	requsest, err := http.NewRequest("POST", "http://auth/signup", bytes.NewBuffer(payload))
	if err != nil {

		return JsonAuthResponse{}, err
	}
	client := &http.Client{}
	response, err := client.Do(requsest)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	defer response.Body.Close()
	var jsResponse JsonAuthResponse
	err = json.NewDecoder(response.Body).Decode(&jsResponse)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	return jsResponse, err

}

func MakeSignInRequest(r *http.Request, email, password, username string) (*JsonAuthResponse, error) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	respch := make(chan ChanAuthStruct)
	go func(email, password, username string) {
		resp, err := CreateUser(email, password, username)
		respch <- ChanAuthStruct{
			JsReq: resp,
			Error: err,
		}
	}(email, password, username)

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("called auth service took to much time")
		case resp := <-respch:
			return &resp.JsReq, resp.Error
		}
	}
}

func (app *Config) HandleGetSgForm(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	w.WriteHeader(http.StatusOK)
	templ.ExecuteTemplate(w, "sgform", nil)
}

func (app *Config) HandleGetLogForm(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	w.WriteHeader(http.StatusOK)
	templ.ExecuteTemplate(w, "logform", nil)
}

func LogInUser(email, password string) (JsonAuthResponse, error) {
	reqdata := User{
		Email:    email,
		Password: password,
	}
	payload, err := json.Marshal(reqdata)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	requsest, err := http.NewRequest("POST", "http://auth/login", bytes.NewBuffer(payload))
	if err != nil {

		return JsonAuthResponse{}, err
	}
	client := &http.Client{}
	response, err := client.Do(requsest)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	defer response.Body.Close()
	var jsResponse JsonAuthResponse
	err = json.NewDecoder(response.Body).Decode(&jsResponse)
	if err != nil {

		return JsonAuthResponse{}, err
	}
	return jsResponse, err

}

func MakeLogInRequest(r *http.Request, email, password string) (*JsonAuthResponse, error) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	respch := make(chan ChanAuthStruct)
	go func(email, password string) {
		resp, err := LogInUser(email, password)
		respch <- ChanAuthStruct{
			JsReq: resp,
			Error: err,
		}
	}(email, password)

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("called auth service took to much time")
		case resp := <-respch:
			return &resp.JsReq, resp.Error
		}
	}
}

func (app *Config) HandleLogInRequest(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	templ := template.Must(template.ParseFS(templateFS, "templates/main.html.gohtml"))
	var errResp JsonAuthResponse
	errResp.Error = true
	if email == "" || password == "" {
		w.WriteHeader(http.StatusOK)
		errResp.Message = "Invalid credentials!"
		templ.ExecuteTemplate(w, "logform", errResp)
		return
	}
	resp, err := MakeLogInRequest(r, email, password)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		errResp.Message = "Invalid credentials!"
		templ.ExecuteTemplate(w, "logform", errResp)
		return
	}
	if resp.Error {
		w.WriteHeader(http.StatusOK)
		errResp.Message = "Invalid credentials!"
		templ.ExecuteTemplate(w, "logform", errResp)
		return
	}
	w.WriteHeader(http.StatusOK)
	templ.ExecuteTemplate(w, "newindex", resp)
}
