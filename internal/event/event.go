package event

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID        string                 `json:"event_id" bson:"event_id"`
	Type      string                 `json:"type" bson:"event_type"`
	Data      map[string]interface{} `json:"data" bson:"data"`
	Timestamp time.Time              `json:"timestamp" bson:"timestamp"`
	Processed bool                   `json:"processed" bson:"processed"`
	ObjectID  primitive.ObjectID     `json:"-" bson:"_id,omitempty"`
}

type CreateEventRequest struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type CreateEventResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

func NewEvent(eventType string, data map[string]interface{}) *Event {
	return &Event{
		ID:        "evt_" + primitive.NewObjectID().Hex()[:12],
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Processed: false,
	}
}
