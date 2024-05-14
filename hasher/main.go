package main

import (
	"database/sql"
	"encoding/json"
	"hasher/data"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/shomali11/util/xhashes"
)

type Config struct {
}

type JsonRequest struct {
	Hash string `json:"hash" bson:"hash"`
}

const postgresurl = "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=UTC connect_timeout=5"

func main() {
	mux := chi.NewMux()
	app := Config{}
	mux.Use(middleware.Recoverer)
	mux.Post("/hash", app.GetHash)
	srv := &http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	db := ConnectToDb(postgresurl)
	data.NewDb(db)
	data.CreateTable()
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
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

func OpneDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ConnectToDb(dsn string) *sql.DB {
	counts := 0
	for {
		connection, err := OpneDB(dsn)
		if err != nil {
			counts++
			log.Println(err)
		} else {
			log.Println("Connected to database")
			return connection
		}
		if counts == 10 {
			log.Println(err)
			return nil
		}
		log.Println("Baking off for two seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
