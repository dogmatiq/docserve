CREATE SCHEMA IF NOT EXISTS docserve;

CREATE TABLE IF NOT EXISTS docserve.repository (
    id             INT PRIMARY KEY,
    full_name      TEXT NOT NULL,
    commit_hash    TEXT NOT NULL,
    is_stale       BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS repository_stale_idx
ON docserve.repository (is_stale);

CREATE TABLE IF NOT EXISTS docserve.type (
    id            SERIAL PRIMARY KEY,
    package       TEXT NOT NULL,
    name          TEXT NOT NULL,
    needs_removal BOOLEAN NOT NULL DEFAULT FALSE,
    repository_id INT,
    url           TEXT,
    docs          TEXT,

    UNIQUE (package, name),

    CONSTRAINT repository_fkey
        FOREIGN KEY (repository_id)
        REFERENCES docserve.repository (id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS type_repository_idx
ON docserve.type (repository_id);

CREATE TABLE IF NOT EXISTS docserve.application (
    key           TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    type_id       INT NOT NULL,
    is_pointer    BOOLEAN NOT NULL,
    repository_id INT NOT NULL,
    needs_removal BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT repository_fkey
        FOREIGN KEY (repository_id)
        REFERENCES docserve.repository (id)
        ON DELETE CASCADE,

    CONSTRAINT type_fkey
        FOREIGN KEY (type_id)
        REFERENCES docserve.type (id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS application_repository_idx
ON docserve.application (repository_id);

CREATE INDEX IF NOT EXISTS application_type_idx
ON docserve.application (type_id);

CREATE TABLE IF NOT EXISTS docserve.handler (
    key             TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    application_key TEXT NOT NULL,
    handler_type    TEXT NOT NULL,
    type_id         INT NOT NULL,
    is_pointer      BOOLEAN NOT NULL,
    needs_removal   BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT application_fkey
        FOREIGN KEY (application_key)
        REFERENCES docserve.application (key)
        ON DELETE CASCADE,

    CONSTRAINT type_fkey
        FOREIGN KEY (type_id)
        REFERENCES docserve.type (id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS handler_application_idx
ON docserve.handler (application_key);

CREATE INDEX IF NOT EXISTS handler_type_idx
ON docserve.handler (type_id);

CREATE TABLE IF NOT EXISTS docserve.handler_message (
    handler_key   TEXT NOT NULL,
    type_id       INT NOT NULL,
    is_pointer    BOOLEAN NOT NULL,
    role          TEXT NOT NULL,
    is_produced   BOOLEAN NOT NULL DEFAULT FALSE,
    is_consumed   BOOLEAN NOT NULL DEFAULT FALSE,
    needs_removal BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (handler_key, type_id, is_pointer),

    CONSTRAINT handler_fkey
        FOREIGN KEY (handler_key)
        REFERENCES docserve.handler (key)
        ON DELETE CASCADE,

    CONSTRAINT type_fkey
        FOREIGN KEY (type_id)
        REFERENCES docserve.type (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS handler_message_type_idx
ON docserve.handler_message (type_id);

CREATE INDEX IF NOT EXISTS handler_message_produced_idx
ON docserve.handler_message (is_produced, type_id);

CREATE INDEX IF NOT EXISTS handler_message_consumed_idx
ON docserve.handler_message (is_consumed, type_id);
