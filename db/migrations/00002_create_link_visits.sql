-- +goose Up
CREATE TABLE IF NOT EXISTS link_visits (
    id         BIGSERIAL    PRIMARY KEY,
    link_id    BIGINT       NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip         TEXT         NOT NULL DEFAULT '',
    referer    TEXT         NOT NULL DEFAULT '',
    user_agent TEXT         NOT NULL DEFAULT '',
    status     INTEGER      NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_link_visits_link_id ON link_visits(link_id);

-- +goose Down
DROP TABLE IF EXISTS link_visits;
