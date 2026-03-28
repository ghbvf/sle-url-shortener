CREATE TABLE links (
    id         BIGSERIAL PRIMARY KEY,
    code       TEXT NOT NULL UNIQUE,
    url        TEXT NOT NULL,
    clicks     BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_links_code ON links(code);
CREATE INDEX idx_links_expires ON links(expires_at) WHERE expires_at IS NOT NULL;
