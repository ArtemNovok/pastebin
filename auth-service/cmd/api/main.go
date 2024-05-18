package main

import (
	"database/sql"
	"fmt"
	"log"
	"logins/data"
	"net/http"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	webPort     = "80"
	postgresUrl = "host=postgres2 user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=UTC connect_timeout=5"
)

type Config struct{}

func main() {
	db := connectToPostgres(postgresUrl)
	defer db.Close()
	err := data.NewDB(db)
	if err != nil {
		log.Panic("Failed to setup db")
		log.Fatal(err)

	}
	app := Config{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func connectToPostgres(url string) *sql.DB {
	count := 0
	for {
		count++
		db, err := sql.Open("pgx", url)
		if err != nil {
			log.Panic(err)
		}
		err = db.Ping()
		if err == nil {
			log.Println("Connected to db")
			return db
		}
		if count > 10 {
			log.Fatalln("Failed to connect to db")
			return nil
		}
		log.Println("Backing off for 2 seconds...")
		time.Sleep(time.Second * 2)
		continue
	}
}
