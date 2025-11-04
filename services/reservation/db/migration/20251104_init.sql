-- migrate:up
CREATE TABLE Ticket (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservationId UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    zoneNumber INTEGER NOT NULL,
    rowNumber INTEGER NOT NULL,
    colNumber INTEGER NOT NULL,

    FOREIGN KEY (reservationId) REFERENCES Reservation(id)
);

CREATE TABLE Reservation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    userId UUID NOT NULL,
    eventId UUID NOT NULL,
    status TEXT NOT NULL,

    FOREIGN KEY (userId) REFERENCES Users(id),
    FOREIGN KEY (eventId) REFERENCES Events(id)
});

CREATE TABLE ReservationTicket (
    reservationId UUID NOT NULL,
    ticketId UUID NOT NULL,

    PRIMARY KEY (reservationId, ticketId),
    FOREIGN KEY (reservationId) REFERENCES Reservation(id) ON DELETE CASCADE,
    FOREIGN KEY (ticketId) REFERENCES Ticket(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS Tickets;
