package domains

func (r *ReserveDomainImpl) Reserve() error {
	// 0. Check the condition if the requested information is valid, cache check
	// 1. Create the create new id, prepare for unique reservation
	// 2. Prepare for unique reservation
	// 3. Create the reservation
	// 4. Save the reservation to cache, in form userId:reservationId
	// 5. Save the reservation to databse, to track the value
	// 6. Save the seat ticket to cache with event:location:zone_number:row:col
	return nil
}

func (r *ReserveDomainImpl) Cancel() error {
	// 0. Check the condition if the requested information is valid, cache check
	// 1. Delete the reservation from cache, in form userId:reservationId
	// 2. Delete the reservation from databse, to track the value
	// 3. Delete the seat ticket from cache with event:location:zone_number:row:col
	return nil
}

func (r *ReserveDomainImpl) Confirm() error {
	// 0. Check the condition if the requested information is valid, cache check
	// 1. Update the reservation status to confirmed
	// 2. Update the reservation status to confirmed in database
	// 3. Update the seat ticket status to confirmed in cache
	return nil
}