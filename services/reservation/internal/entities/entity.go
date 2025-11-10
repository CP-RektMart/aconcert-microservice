package entities

import "time"

type ReservationStatus string

const (
	Confirmed ReservationStatus = "CONFIRMED"
	Pending   ReservationStatus = "PENDING"
	Cancelled ReservationStatus = "CANCELLED"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatPending   SeatStatus = "PENDING"
	SeatReserved  SeatStatus = "RESERVED"
)

type Reservation struct {
	ID         string            `json:"id"`
	UserID     string            `json:"user_id"`
	EventID    string            `json:"event_id"`
	Tickets    []Ticket          `json:"tickets"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Status     ReservationStatus `json:"status"`
	TotalPrice float64           `json:"total_price"`
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

type ConfirmedNotiReservation struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
}

type CancelledNotiReservation struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}
