package pubsub

// SeatUpdate represents a seat status change message from Redis Pub/Sub
type SeatUpdate struct {
	EventID    string `json:"eventId"`
	ZoneNumber int32  `json:"zoneNumber"`
	Row        int32  `json:"row"`
	Column     int32  `json:"column"`
	Status     string `json:"status"` // "AVAILABLE", "PENDING", "RESERVED"
	Timestamp  int64  `json:"timestamp"`
}

// IsValid checks if the seat update has valid data
func (su *SeatUpdate) IsValid() bool {
	return su.EventID != "" &&
		su.ZoneNumber > 0 &&
		su.Row > 0 &&
		su.Column > 0 &&
		(su.Status == "AVAILABLE" ||
			su.Status == "PENDING" ||
			su.Status == "RESERVED")
}
