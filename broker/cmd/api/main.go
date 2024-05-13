package main

import (
	"broker/data"
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct{}

const mongoUrl = "mongodb://host.docker.internal:27017"

func main() {
	app := Config{}
	server := &http.Server{
		Addr:    ":8000",
		Handler: app.routes(),
	}
	client := ConnectTOMongoDB(mongoUrl)
	data.NewClient(client)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func ConnectTOMongoDB(url string) *mongo.Client {
	count := 0
	for {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))
		if err == nil {
			log.Println("Successfully connected to mongodb")
			return client
		}
		if count >= 8 {
			log.Panic(err)
			log.Panic("Failed to connect to Mongodb")
			break
		}
		count++
		log.Println("Backing off for 2 seconds...")
		time.Sleep(time.Second * 2)
	}
	return nil
}
