CREATE TABLE IF NOT EXISTS messages
(
    id INTEGER PRIMARY KEY,
    content TEXT NOT NULL,
    uid INTEGER,
    type INTEGER,
    datetime TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS types
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE
);

INSERT INTO types (name)
VALUES ('text'),
       ('image'),
       ('file')