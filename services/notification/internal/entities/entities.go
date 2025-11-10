package entities

import (
	"encoding/json"
)

type MessageType string

const (
	MessageTypeEventCreated         MessageType = "event.created"
	MessageTypeEventUpdated         MessageType = "event.updated"
	MessageTypeReservationConfirmed MessageType = "reservation.confirmed"
	MessageTypeReservationCancelled MessageType = "reservation.cancelled"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Event struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	LocationID  string   `json:"locationId"`
	Artist      []string `json:"artist"`
	EventDate   string   `json:"eventDate"`
	Thumbnail   string   `json:"thumbnail"`
	Images      []string `json:"images"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type ConfirmedNotiReservation struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
}

type CancelledNotiReservation struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
}
