CREATE TABLE review
(
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description     TEXT NOT NULL,
    author_username UUID REFERENCES employee (id) ON DELETE CASCADE,
    created_at      TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);