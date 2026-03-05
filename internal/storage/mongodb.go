package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBStore struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewMongoDBStore(ctx context.Context, uri, dbName string) (*MongoDBStore, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database(dbName)
	collection := db.Collection("events")

	fmt.Println("Connected to MongoDB")

	return &MongoDBStore{
		client:     client,
		database:   db,
		collection: collection,
	}, nil
}

func (s *MongoDBStore) InsertMany(ctx context.Context, events []*event.Event) error {
	docs := make([]interface{}, len(events))
	for i, evt := range events {
		docs[i] = evt
	}

	_, err := s.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to insert events: %w", err)
	}
	return nil
}

func (s *MongoDBStore) FindByID(ctx context.Context, eventID string) (*event.Event, error) {
	var evt event.Event
	err := s.collection.FindOne(ctx, bson.M{"event_id": eventID}).Decode(&evt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find event: %w", err)
	}
	return &evt, nil
}

func (s *MongoDBStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func (s *MongoDBStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx, nil)
}
