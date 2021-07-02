CREATE SCHEMA IF NOT EXISTS docserve;

CREATE TABLE IF NOT EXISTS docserve.repository (
    id          INT PRIMARY KEY,
    full_name   TEXT NOT NULL,
    commit_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS docserve.application (
    key           TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    type_name     TEXT NOT NULL,
    repository_id INT NOT NULL,

    CONSTRAINT repository_fkey
        FOREIGN KEY (repository_id)
        REFERENCES docserve.repository (id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS docserve.handler (
    key             TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    application_key TEXT NOT NULL,
    handler_type    TEXT NOT NULL,
    type_name       TEXT NOT NULL,

    CONSTRAINT application_fkey
        FOREIGN KEY (application_key)
        REFERENCES docserve.application (key)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS docserve.handler_message (
    handler_key TEXT NOT NULL,
    type_name   TEXT NOT NULL,
    role        TEXT NOT NULL,
    produced    BOOLEAN NOT NULL DEFAULT False,
    consumed    BOOLEAN NOT NULL DEFAULT False,

    PRIMARY KEY (handler_key, type_name),
    CONSTRAINT handler_fkey
        FOREIGN KEY (handler_key)
        REFERENCES docserve.handler (key)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS handler_message_type_name_idx
    ON docserve.handler_message (type_name);
