CREATE SCHEMA docserve;

CREATE TABLE docserve.application (
    key      TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    type_name TEXT NOT NULL
);

CREATE TABLE docserve.handler (
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

CREATE TABLE docserve.handler_message (
    handler_key TEXT NOT NULL,
    type_name   TEXT NOT NULL,
    role        TEXT NOT NULL,
    produced    BOOLEAN,
    consumed    BOOLEAN,

    PRIMARY KEY (handler_key, type_name),
    CONSTRAINT handler_fkey
        FOREIGN KEY (handler_key)
        REFERENCES docserve.handler (key)
        ON DELETE CASCADE
);
