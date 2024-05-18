package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"logins/data"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

const keystr = "secret"

func (app *Config) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	var user data.User

	log.Println("Handled sign up")
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}
	if user.Email == "" || user.Password == "" || user.UserName == "" {
		WriteError(w, http.StatusBadRequest, "Empty user or password or username fields")
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate hash from password")
		return
	}
	user.Password = string(hashedPass)
	err = user.SaveUser()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to save new user in db")
		return
	}
	payload := JsonResponse{
		Error:   false,
		Message: "User was successfully created!!",
	}
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(payload)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to encode payload")
		return
	}
}

func (app *Config) HandlerLogIn(w http.ResponseWriter, r *http.Request) {
	var user data.User
	log.Println("Handled login")
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("1")
		WriteError(w, http.StatusBadRequest, "Failed to decode request body")
		return
	}
	if user.Email == "" || user.Password == "" {
		log.Println("2")
		WriteError(w, http.StatusBadRequest, "Empty email or password fields")
		return
	}
	new_user, err := user.GetUserByEmail()
	if err == sql.ErrNoRows {
		log.Println("3")
		WriteError(w, http.StatusNotFound, "User doesn't exists")
		return
	}
	if err != nil {
		log.Println("4")
		WriteError(w, http.StatusInternalServerError, "Failed find user")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(new_user.Password), []byte(user.Password))
	if err != nil {
		log.Println("5")
		WriteError(w, http.StatusBadRequest, "Incorrect password")
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  new_user.Id,
		"name": new_user.UserName,
		"exp":  time.Now().Add(time.Second * 7200).Unix(),
	})
	tokenstr, err := token.SignedString([]byte(keystr))
	if err != nil {
		log.Println("6")
		WriteError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}
	payload := JsonResponse{
		Error:   false,
		Message: "Token was successfully created",
		Token:   tokenstr,
	}
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(payload)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to encode payload")
		return
	}
}
