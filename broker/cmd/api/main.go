package main

import (
	"broker/data"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct{}

const (
	mongoUrl = "mongodb://mongo"
	webPort  = "8000"
	redisAdr = "redis:6379"
)

func main() {
	app := Config{}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	client := ConnectTOMongoDB(mongoUrl)

	rediscl := ConnectTORedis(redisAdr)

	defer rediscl.Close()

	data.NewRedisClient(rediscl)

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
		err = client.Ping(context.TODO(), nil)
		if err == nil {
			log.Println("successfully connected to mongodb")
			return client
		}
		if count >= 8 {
			log.Panic(err)
			log.Panic("failed to connect to Mongodb")
			break
		}
		count++
		log.Println("backing off for 2 seconds...")
		time.Sleep(time.Second * 2)
	}
	return nil
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
