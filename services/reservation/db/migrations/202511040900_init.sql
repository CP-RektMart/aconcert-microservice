-- migrate:up
CREATE TABLE Reservation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE Ticket (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    zone_number INTEGER NOT NULL,
    row_number INTEGER NOT NULL,
    col_number INTEGER NOT NULL,

    FOREIGN KEY (reservation_id) REFERENCES Reservation(id)
);

CREATE TABLE ReservationTicket (
    reservation_id UUID NOT NULL,
    ticket_id UUID NOT NULL,

    PRIMARY KEY (reservation_id, ticket_id),
    FOREIGN KEY (reservation_id) REFERENCES Reservation(id) ON DELETE CASCADE,
    FOREIGN KEY (ticket_id) REFERENCES Ticket(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS Tickets;
