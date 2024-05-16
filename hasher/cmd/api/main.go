package main

import (
	"database/sql"
	"hasher/data"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Config struct{}

const postgresurl = "host=postgres user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=UTC connect_timeout=5"

func main() {
	app := Config{}
	srv := &http.Server{
		Addr:    ":80",
		Handler: app.routes(),
	}
	db := ConnectToDb(postgresurl)
	defer db.Close()
	data.NewDb(db)
	data.CreateTable()
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
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
