-- +migrate Up
CREATE TABLE links(
    id SERIAL PRIMARY KEY,
    original_link TEXT NOT NULL UNIQUE,
    shortened_link CHAR(10) NOT NULL UNIQUE
)

CREATE INDEX on links(original_link) USING HASH
CREATE INDEX on links(shortened_link) USING HASH