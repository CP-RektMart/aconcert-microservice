package dto

// SeatDTO represents a seat in a reservation
type SeatDTO struct {
	ZoneNumber int32   `json:"zoneNumber" validate:"required"`
	Row        int32   `json:"row" validate:"required"`
	Column     int32   `json:"column" validate:"required"`
	Price      float64 `json:"price,omitempty"`
}

// CreateReservationRequest is the request body for creating a reservation
type CreateReservationRequest struct {
	EventID    string    `json:"eventId" validate:"required"`
	TotalPrice float64   `json:"totalPrice" validate:"required"`
	Seats      []SeatDTO `json:"seats" validate:"required,min=1"`
}

// CreateReservationResponse is the response for creating a reservation
type CreateReservationResponse struct {
	ID string `json:"id" validate:"required"`
}

// DeleteReservationRequest is the request for deleting a reservation
type DeleteReservationRequest struct {
	ID string `params:"id" validate:"required"`
}

// DeleteReservationResponse is the response for deleting a reservation
type DeleteReservationResponse struct {
	ID string `json:"id" validate:"required"`
}

// GetReservationRequest is the request for getting a reservation
type GetReservationRequest struct {
	ID string `params:"id" validate:"required"`
}

// GetReservationResponse is the response for getting a reservation
type GetReservationResponse struct {
	ID                 string    `json:"id" validate:"required"`
	UserID             string    `json:"userId" validate:"required"`
	EventID            string    `json:"eventId" validate:"required"`
	TotalPrice         float64   `json:"totalPrice" validate:"required"`
	Seats              []SeatDTO `json:"seats" validate:"required"`
	StripeClientSecret string    `json:"stripeClientSecret" validate:"required"`
	TimeLeft           float64   `json:"timeLeft" validate:"required"`
}

// ListReservationRequest is the request for listing reservations
type ListReservationRequest struct {
	UserID string `query:"userId" validate:"required"`
}

// ListReservationResponse is the response for listing reservations
type ListReservationResponse struct {
	Reservations []ReservationDTO `json:"reservations" validate:"required"`
}

// ReservationDTO represents a reservation in the list
type ReservationDTO struct {
	ID         string    `json:"id,omitempty"`
	UserID     string    `json:"userId" validate:"required"`
	EventID    string    `json:"eventId" validate:"required"`
	TotalPrice float64   `json:"totalPrice"`
	Seats      []SeatDTO `json:"seats"`
}

// ConfirmReservationRequest is the request for confirming a reservation
type ConfirmReservationRequest struct {
	ID string `params:"id" validate:"required" swaggerignore:"true"`
}

// ConfirmReservationResponse is the response for confirming a reservation
type ConfirmReservationResponse struct {
	ID      string `json:"id" validate:"required"`
	Success bool   `json:"success" validate:"required"`
	Message string `json:"message" validate:"required"`
}
