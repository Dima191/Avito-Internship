CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
);

CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
);

CREATE TABLE tender (
    id               UUID PRIMARY KEY       DEFAULT uuid_generate_v4(),
    name             VARCHAR(100)  NOT NULL,
    description      TEXT          NOT NULL,
    service_type     service_type  NOT NULL,
    status           tender_status NOT NULL DEFAULT 'Created',
    organization_id  UUID REFERENCES organization (id) ON DELETE CASCADE,
    creator_username VARCHAR(100)  NOT NULL,
    version       INT                    DEFAULT 1,
    created_at       TIMESTAMP              DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tender_history (
    id               UUID,
    name             VARCHAR(100),
    description      TEXT,
    service_type     service_type,
    status           tender_status,
    organization_id  UUID,
    creator_username VARCHAR(100),
    version       INT,
    created_at       TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, version)
);

CREATE OR REPLACE FUNCTION update_tender_version()
RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO tender_history (id, name, description, service_type, status, organization_id, creator_username, version, created_at, updated_at)
        VALUES (OLD.id, OLD.name, OLD.description, OLD.service_type, OLD.status, OLD.organization_id, OLD.creator_username, OLD.version, OLD.created_at, CURRENT_TIMESTAMP);

        NEW.version := OLD.version + 1;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_tender_version BEFORE UPDATE ON tender
FOR EACH ROW EXECUTE FUNCTION update_tender_version();