-- +goose Up
CREATE TABLE IF NOT EXISTS links (
    id           BIGSERIAL    PRIMARY KEY,
    original_url TEXT         NOT NULL,
    short_name   VARCHAR(32)  NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS links;
