package storage

import (
	"context"
	"log"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//
//type Emotion struct {
//	Name  string `bson:"name"`
//	Scale int    `bson:"scale"`
//}
//
//type Thought struct {
//	Description string `bson:"description"`
//}
//
//type Action struct {
//	Description string `bson:"description"`
//}
//
//type Trigger struct {
//	Description string `bson:"description"`
//}

type SMEREntry struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	UserID       int64                `bson:"user_id"`
	CreatedTime  time.Time            `bson:"created_time"`
	UpdatedTime  time.Time            `bson:"updated_time"`
	Emotions     []*entity.Emotion    `bson:"emotions"`
	Thoughts     []*entity.Thought    `bson:"thoughts"`
	Action       *entity.Action       `bson:"action"`
	Trigger      *entity.Trigger      `bson:"trigger"`
	Unstructured *entity.Unstructured `bson:"unstructured"`
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

func (adapter *SMERStorage) Save(ctx context.Context, entry *entity.SMEREntry) (*entity.SMEREntry, error) {
	now := time.Now()
	if entry.CreatedTime.IsZero() {
		entry.CreatedTime = now
	}
	entry.UpdatedTime = now

	if entry.ID == "" {
		entry.ID = primitive.NewObjectID().Hex()
	}
	id_, err := primitive.ObjectIDFromHex(entry.ID)
	if err != nil {
		return nil, err
	}

	db_entry := SMEREntry{
		ID:           id_,
		UserID:       entry.UserID,
		CreatedTime:  entry.CreatedTime,
		UpdatedTime:  entry.UpdatedTime,
		Trigger:      entry.Trigger,
		Emotions:     entry.Emotions,
		Thoughts:     entry.Thoughts,
		Unstructured: entry.Unstructured,
	}

	filter := bson.M{"_id": id_}
	update := bson.M{"$set": db_entry}
	opts := options.Update().SetUpsert(true)

	_, err = adapter.smerCollection.UpdateOne(ctx, filter, update, opts)
	return entry, err
}

func (adapter *SMERStorage) GetByID(id string) (*entity.SMEREntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": id_}
	res := adapter.smerCollection.FindOne(ctx, filter)
	var entry SMEREntry
	err = res.Decode(&entry)
	return &entity.SMEREntry{
		ID:          entry.ID.Hex(),
		UserID:      entry.UserID,
		CreatedTime: entry.CreatedTime,
		UpdatedTime: entry.UpdatedTime,
		Trigger:     entry.Trigger,
		Emotions:    entry.Emotions,
		Thoughts:    entry.Thoughts,
		//Action:      entry.Action,
	}, err
}

func (adapter *SMERStorage) GetByUserID(ctx context.Context, id int64, startDate time.Time, endDate time.Time) ([]*entity.SMEREntry, error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"created_time", bson.D{{"$gte", startDate}}}},
				bson.D{{"created_time", bson.D{{"$lte", endDate}}}},
			}},
	}

	cursor, err := adapter.smerCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*SMEREntry
	for cursor.Next(ctx) {
		var entry SMEREntry
		if err := cursor.Decode(&entry); err != nil {
			log.Println("Ошибка декодирования:", err)
			continue
		}
		entries = append(entries, &entry)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	res := make([]*entity.SMEREntry, len(entries))
	for i, entry := range entries {
		res[i] = &entity.SMEREntry{
			ID:           entry.ID.Hex(),
			UserID:       entry.UserID,
			CreatedTime:  entry.CreatedTime,
			UpdatedTime:  entry.UpdatedTime,
			Trigger:      entry.Trigger,
			Emotions:     entry.Emotions,
			Thoughts:     entry.Thoughts,
			Unstructured: entry.Unstructured,
		}
	}

	return res, nil
}

func (adapter *SMERStorage) DeleteByID(ctx context.Context, id string) error {
	id_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": id_}
	_, err = adapter.smerCollection.DeleteOne(ctx, filter)
	return err
}

func (adapter *SMERStorage) Close(ctx context.Context) error {
	return adapter.client.Disconnect(ctx)
}
