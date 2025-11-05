package entities

import "time"

// We assumed the reservation collect in DB is already confirmed and paid
type Reservation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	Tickets   []Ticket  `json:"tickets"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Ticket struct {
	ID            string    `json:"id"`
	EventID       string    `json:"event_id"`
	ReservationID string    `json:"reservation_id"`
	Price         float64   `json:"price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Event struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	LocationID string    `json:"location_id"`
	Capacity   int       `json:"capacity"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
