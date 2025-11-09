package entities

import "github.com/google/uuid"

type EventType string

const (
	EventTypeEstimation      EventType = "ESTIMATION"
	EventTypeEstimationError EventType = "ESTIMATION_ERROR"
	EventTypeAdminMessage    EventType = "ADMIN_MESSAGE"
)

type EventData struct {
	EventID   uuid.UUID `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	EventType EventType `json:"eventType"`
	Data      any       `json:"data"`
}
