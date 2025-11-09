-- migrate:up

ALTER TABLE Ticket ADD COLUMN event_id UUID;

UPDATE Ticket t
SET event_id = r.event_id
FROM ReservationTicket rt
JOIN Reservation r ON rt.reservation_id = r.id
WHERE t.id = rt.ticket_id;

ALTER TABLE Ticket ALTER COLUMN event_id SET NOT NULL;

-- migrate:down
ALTER TABLE Ticket DROP COLUMN event_id;
