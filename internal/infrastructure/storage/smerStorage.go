package storage

import (
	"context"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SMEREmotion struct {
	EmotionID int64  `bson:"emotion_id"`
	Name      string `bson:"name"`
	SmerID    int64  `bson:"smer_id"`
	Scale     int    `bson:"scale"`
}

type Thought struct {
	ID          int64     `bson:"id"`
	Description string    `bson:"description"`
	SmerID      int64     `bson:"smer_id"`
	CreatedTime time.Time `bson:"created_time"`
	UpdatedTime time.Time `bson:"updated_time"`
}

type SMEREntry struct {
	ID          string        `bson:"_id,omitempty"`
	UserID      int64         `bson:"user_id"`
	CreatedTime time.Time     `bson:"created_time"`
	UpdatedTime time.Time     `bson:"updated_time"`
	Emotions    []SMEREmotion `bson:"emotions"`
	Thoughts    []Thought     `bson:"thoughts"`
}

type SMERStorage struct {
	client         *mongo.Client
	smerCollection *mongo.Collection
}

func NewSMERStorage(uri, dbName string) (*SMERStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection("smer_entry")

	return &SMERStorage{
		client:         client,
		smerCollection: collection,
	}, nil
}

func (adapter *SMERStorage) Save(ctx context.Context, entry *entity.SMEREntry) error {
	now := time.Now()
	if entry.CreatedTime.IsZero() {
		entry.CreatedTime = now
	}
	entry.UpdatedTime = now

	filter := bson.M{"_id": primitive.NewObjectID()}
	update := bson.M{"$set": entry}
	opts := options.Update().SetUpsert(true)

	_, err := adapter.smerCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (adapter *SMERStorage) GetByUserID(ctx context.Context, id int64) ([]*entity.SMEREntry, error) {
	filter := bson.M{}
	//filter := bson.M{"user_id": id}  todo: find out why this filter gives empty result

	cursor, err := adapter.smerCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*entity.SMEREntry
	for cursor.Next(ctx) {
		var entry entity.SMEREntry
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Закрытие соединения

func (adapter *SMERStorage) Close(ctx context.Context) error {
	return adapter.client.Disconnect(ctx)
}
