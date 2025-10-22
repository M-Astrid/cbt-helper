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
	ID           string               `bson:"_id,omitempty"`
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

func (adapter *SMERStorage) Save(ctx context.Context, entry *entity.SMEREntry) error {
	now := time.Now()
	if entry.CreatedTime.IsZero() {
		entry.CreatedTime = now
	}
	entry.UpdatedTime = now

	id_, err := primitive.ObjectIDFromHex(entry.ID)

	db_entry := SMEREntry{
		UserID:       entry.UserID,
		CreatedTime:  entry.CreatedTime,
		UpdatedTime:  entry.UpdatedTime,
		Trigger:      entry.Trigger,
		Emotions:     entry.Emotions,
		Thoughts:     entry.Thoughts,
		Unstructured: entry.Unstructured,
	}
	//for _, em := range entry.Emotions {
	//	db_entry.Emotions = append(db_entry.Emotions, Emotion{em.Name, em.Scale})
	//}
	//for _, th := range entry.Thoughts {
	//	db_entry.Thoughts = append(db_entry.Thoughts, Thought{th.Description})
	//}
	//if entry.Action != nil {
	//	db_entry.Action = &Action{entry.Action.Description}
	//}

	filter := bson.M{"_id": id_}
	update := bson.M{"$set": db_entry}
	opts := options.Update().SetUpsert(true)

	_, err = adapter.smerCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (adapter *SMERStorage) GetByID(ctx context.Context, id string) (*entity.SMEREntry, error) {
	id_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": id_}
	res := adapter.smerCollection.FindOne(ctx, filter)
	var entry SMEREntry
	err = res.Decode(&entry)
	return &entity.SMEREntry{
		ID:          entry.ID,
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
			ID:           entry.ID,
			UserID:       entry.UserID,
			CreatedTime:  entry.CreatedTime,
			UpdatedTime:  entry.UpdatedTime,
			Trigger:      entry.Trigger,
			Emotions:     entry.Emotions,
			Thoughts:     entry.Thoughts,
			Unstructured: entry.Unstructured,
			//Action:      entry.Action,
		}
		//for _, em := range entry.Emotions {
		//	res[i].Emotions = append(res[i].Emotions, &entity.Emotion{
		//		Name:  em.Name,
		//		Scale: em.Scale,
		//	})
		//}
		//for _, th := range entry.Thoughts {
		//	res[i].Thoughts = append(res[i].Thoughts, &entity.Thought{
		//		Description: th.Description,
		//	})
		//}
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
