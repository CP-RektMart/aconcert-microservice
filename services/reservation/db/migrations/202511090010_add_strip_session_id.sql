-- migrate:up
BEGIN;

ALTER TABLE Reservation ADD COLUMN stripe_session_id TEXT NULL;

UPDATE Reservation SET stripe_session_id = 'test-demo';

ALTER TABLE Reservation
ALTER COLUMN stripe_session_id SET NOT NULL;

COMMIT;
END;

-- migrate:down
ALTER TABLE Reservation DROP COLUMN stripe_session_id;
