CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Canceled'
);

CREATE TYPE author_type AS ENUM (
    'Organization',
    'User'
);

CREATE TABLE bid (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL,
    status      bid_status       DEFAULT 'Created',
    tender_id   UUID REFERENCES tender (id) ON DELETE CASCADE,
    author_type author_type  NOT NULL,
    author_id   UUID,
    version     INT              DEFAULT 1,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE bid_history (
    id          UUID,
    name        VARCHAR(100),
    description TEXT,
    status      bid_status,
    tender_id   UUID,
    author_type author_type,
    author_id   UUID,
    version     INT,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, version)
);

CREATE OR REPLACE FUNCTION update_bid_version()
RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO bid_history (id, name, description, status, tender_id, author_type, author_id, version, created_at, updated_at)
        VALUES (OLD.id, OLD.name, OLD.description, OLD.status, OLD.tender_id, OLD.author_type, OLD.author_id, OLD.version, OLD.created_at, CURRENT_TIMESTAMP);

        NEW.version := OLD.version + 1;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_bid_version BEFORE UPDATE ON bid
FOR EACH ROW EXECUTE FUNCTION update_bid_version();