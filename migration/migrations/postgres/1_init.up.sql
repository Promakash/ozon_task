-- +migrate Up
CREATE TABLE links(
    id SERIAL PRIMARY KEY,
    original_link TEXT NOT NULL UNIQUE,
    shortened_link CHAR(10) NOT NULL UNIQUE
);

CREATE INDEX idx_links_original_link ON links USING HASH (original_link);
CREATE INDEX idx_links_shortened_link ON links USING HASH (shortened_link);
