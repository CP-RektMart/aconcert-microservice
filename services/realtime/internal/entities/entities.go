package entities

import "github.com/google/uuid"

type EventData struct {
	EventID   uuid.UUID `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	EventType string    `json:"eventType"`
	Data      any       `json:"data"`
}
