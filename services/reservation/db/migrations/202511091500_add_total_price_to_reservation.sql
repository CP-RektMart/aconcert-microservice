-- migrate:up
ALTER TABLE Reservation ADD COLUMN total_price FLOAT NULL;

UPDATE Reservation SET total_price = 100;

ALTER TABLE Reservation
ALTER COLUMN total_price SET NOT NULL;

-- migrate:down
ALTER TABLE Reservation DROP COLUMN total_price IF EXISTS;
