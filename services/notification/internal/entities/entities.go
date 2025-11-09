package entities

type MessageType string

const (
	MessageTypeEventCreated         MessageType = "event.created"
	MessageTypeEventUpdated         MessageType = "event.updated"
	MessageTypeReservationConfirmed MessageType = "reservation.confirmed"
	MessageTypeReservationCancelled MessageType = "reservation.cancelled"
)

type Message struct {
	Type MessageType `json:"type"`
	Data any         `json:"data"`
}
