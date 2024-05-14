package data

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func NewClient(cl *mongo.Client) {
	client = cl
	index := mongo.IndexModel{
		Keys:    bson.D{{"TimeToLive", 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	coll := client.Database("mes").Collection("mes")
	name, err := coll.Indexes().CreateOne(context.TODO(), index)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Index %s was created", name)
}

var rediscl *redis.Client

func NewRedisClient(cl *redis.Client) {
	rediscl = cl
}

type Message struct {
	ID   string `json:"id,omitempty" bson:"_id"`
	Text string `json:"text" bson:"text"`
	Hash string `json:"hash,omitempty" bson:"hash,omitempty"`
	HTL  int64  `json:"htl,omitempty" bson:"htl"`
}

func (m *Message) InsertMes() error {
	coll := client.Database("mes").Collection("mes")
	_, err := coll.InsertOne(context.TODO(), bson.D{{"text", m.Text}, {"hash", m.Hash}, {"TimeToLive", time.Now().UTC().Add(time.Second * time.Duration(m.HTL))}})
	if err != nil {
		return err
	}
	return nil
}

func FindMesByHash(hash string) (Message, error) {
	coll := client.Database("mes").Collection("mes")
	var mes Message
	err := coll.FindOne(context.TODO(), bson.D{{"hash", hash}}).Decode(&mes)
	if err != nil {
		return Message{}, err
	}
	return mes, nil
}

func GetByKey(hash string) (Message, error) {
	ctx := context.Background()
	var mes Message
	res, err := rediscl.Get(ctx, hash).Result()
	if err != nil {
		return Message{}, err
	}
	mes.Text = res
	return mes, nil
}

func (m *Message) SetMes() error {
	ctx := context.Background()
	err := rediscl.Set(ctx, m.Hash, m.Text, time.Second*20).Err()
	if err != nil {
		return err
	}
	return nil
}