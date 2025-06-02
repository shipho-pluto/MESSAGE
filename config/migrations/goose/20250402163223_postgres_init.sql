-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS messages
(
    id INTEGER PRIMARY KEY,
    content TEXT NOT NULL,
    uid INTEGER
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
