package data

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func NewClient(cl *mongo.Client) {
	client = cl
}

type Message struct {
	ID   string `json:"id,omitempty" bson:"_id"`
	Text string `json:"text" bson:"text"`
	Hash string `json:"hash,omitempty" bson:"hash,omitempty"`
	HTL  int64  `json:"htl" bson:"htl"`
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
