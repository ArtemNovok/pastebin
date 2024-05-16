package main

import (
	"context"
	"database/sql"
	"hasher/data"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/redis/go-redis/v9"
)

type Config struct{}

const (
	postgResurl = "host=postgres user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=UTC connect_timeout=5"

	redisAddr = "redis2:6379"
)

func main() {
	app := Config{}
	srv := &http.Server{
		Addr:    ":80",
		Handler: app.routes(),
	}
	db := ConnectToDb(postgResurl)
	defer db.Close()
	data.NewDb(db)
	data.CreateTable()
	rediscl := ConnectTORedis(redisAddr)
	defer rediscl.Close()
	data.NewRedisClient(rediscl)
	data.GenerateHashes()
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func ConnectTORedis(adr string) *redis.Client {
	count := 0
	for {
		rediscl := redis.NewClient(&redis.Options{
			Addr:     adr,
			Password: "",
			DB:       0,
		})
		err := rediscl.Ping(context.TODO()).Err()
		if err == nil {
			log.Println("successfully connected to redis")
			return rediscl
		}
		if count >= 8 {
			log.Panic(err)
			log.Panic("failed to connect to redis")
			break
		}
		count++
		log.Println("backing off for 2 seconds...")
		time.Sleep(time.Second * 2)
	}
	return nil
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
