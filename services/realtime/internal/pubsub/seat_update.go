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

// SeatUpdateBatch represents a batch of seat updates
type SeatUpdateBatch struct {
	Type    string       `json:"type"`    // Must be "batch"
	Updates []SeatUpdate `json:"updates"` // Array of seat updates
}

// IsValid checks if the batch message is valid
func (sub *SeatUpdateBatch) IsValid() bool {
	if sub.Type != "batch" || len(sub.Updates) == 0 {
		return false
	}

	// Validate each update in the batch
	for _, update := range sub.Updates {
		if !update.IsValid() {
			return false
		}
	}

	return true
}
