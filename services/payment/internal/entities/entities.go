package entities

type Seat struct {
	ZoneNumber int
	Price      float64
	Row        int
	Column     int
}

type Reservation struct {
	ID         string
	UserID     string
	EventID    string
	TotalPrice float64
	Seats      []Seat
}
