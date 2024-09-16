CREATE TYPE possible_decisions AS ENUM (
    'Approved',
    'Rejected'
);

CREATE TABLE decision
(
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tender_author_id UUID REFERENCES employee (id) ON DELETE CASCADE,
    tender_id        UUID REFERENCES tender (id) ON DELETE CASCADE,
    bid_id      UUID REFERENCES bid (id) ON DELETE CASCADE,
    decision         possible_decisions NOT NULL,

    CONSTRAINT unique_tender_author UNIQUE (tender_id, tender_author_id)
);
