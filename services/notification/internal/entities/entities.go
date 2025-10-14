package entities

import "github.com/google/uuid"

type EventType string

const (
	EventTypeMessageBroadcast EventType = "MESSAGE_BROADCAST"
	EventTypeMessageSent      EventType = "MESSAGE_SENT"
	EventTypeMessageError     EventType = "MESSAGE_ERROR"
)

type EventData struct {
	EventID   uuid.UUID `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	EventType EventType `json:"eventType"`
	Data      any       `json:"data"`
}
