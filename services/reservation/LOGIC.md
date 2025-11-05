# Reservation Service Logic

## Overview

Temporary reservation system with 5-minute TTL. Tickets created only after payment confirmation.

## Status Flow

```
pending → confirmed (payment success)
pending → cancelled (timeout/manual cancel)
```

## Core Operations

### 1. CreateReservation (Reserve)

**Purpose**: Hold seats temporarily before payment

**Steps**:

1. Validate request (userID, eventID, seats required)
2. Check seat availability in Redis cache
3. Create reservation in DB (status: pending)
4. Cache temp reservation (TTL: 5min)
5. Cache seat information (TTL: 5min)
6. Mark seats as reserved in cache (TTL: 5min)

**Rollback on failure**: Delete cache + soft delete DB reservation

**Note**: No tickets created at this stage

---

### 2. ConfirmReservation (Payment Success)

**Purpose**: Finalize reservation and create tickets

**Steps**:

1. Verify reservation exists and not expired
2. Retrieve cached seat information
3. **Create tickets in DB** (first time)
4. Update reservation status to "confirmed"
5. Delete temp cache
6. Release seat locks in cache
7. Delete cached seats

**Condition**: Must be called within 5 minutes of reservation (with 30s safety buffer)

---

### 3. DeleteReservation (Cancel)

**Purpose**: Cancel pending reservation

**Steps**:

1. Fetch reservation from DB
2. Verify not expired via cache
3. Delete temp cache
4. Soft delete reservation in DB
5. Release seat locks
6. Delete cached seats

**Note**: Only pending reservations can be cancelled (with 30s safety buffer)

---

### 4. GetReservation (Query)

**Purpose**: Fetch reservation details

**Returns**:

- Reservation info (userID, eventID)
- Associated tickets (only if confirmed)
- Seat information

---

## Data Storage

### PostgreSQL (Persistent)

- **Reservation**: id, userID, eventID, status, timestamps
- **Ticket**: id, reservationID, zoneNumber, rowNumber, colNumber
- Soft delete with `deleted_at`

### Redis (Temporary)

- `reservation:temp:{userID}:{reservationID}` → TTL tracking
- `reservation:seats:{reservationID}` → Seat info (JSON)
- `seat:{eventID}:{zone}:{row}:{col}` → Seat lock

**TTL**: 5 minutes for all temp data

---

## Key Conditions

### Seat Availability

- Must check cache before reservation
- Reject if seat already locked
- TTL auto-releases expired locks

### Ticket Creation

- **Only on ConfirmReservation**
- Requires cached seat data
- Cannot create if reservation expired

### Expiration Handling

- Auto-expire via Redis TTL
- Manual check on operations
- Expired = cannot confirm/cancel
- Safety buffer prevents operations if <30s remaining

### Safety Buffer (30 seconds)

- Prevents race conditions from network latency
- Blocks cancel if <30s left
- Blocks confirm if <30s left
- User must create new reservation if too close to expiry

### Rollback Strategy

- Delete temp cache
- Delete cached seats
- Soft delete DB reservation
- Release all seat locks

---

## Error Types

- `BadRequest` (400): Validation, seat taken
- `NotFound` (404): Reservation not found/expired
- `Internal` (500): DB/cache failures

---

## Constants

```go
ReservationTTL  = 5 * time.Minute
SafetyBuffer    = 30 * time.Second
StatusPending   = "pending"
StatusConfirmed = "confirmed"
StatusCancelled = "cancelled"
```
