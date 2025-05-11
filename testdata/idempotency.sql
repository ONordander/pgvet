CREATE TABLE pgvet (id text PRIMARY KEY);

-- pgvet_nolint:missing-if-not-exists
CREATE TABLE pgvet (id text PRIMARY KEY);

CREATE TABLE IF NOT EXISTS pgvet (id text PRIMARY KEY);

CREATE INDEX CONCURRENTLY pgvet_idx ON pgvet(id);

-- pgvet_nolint:missing-if-not-exists
CREATE INDEX CONCURRENTLY pgvet_idx ON pgvet(id);

-- index without name should succeed
CREATE INDEX CONCURRENTLY ON pgvet(id);

ALTER TABLE pgvet ADD COLUMN value text;

-- pgvet_nolint:missing-if-not-exists
ALTER TABLE pgvet ADD COLUMN value text;
