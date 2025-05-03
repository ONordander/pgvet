CREATE TABLE pgcheck (id text PRIMARY KEY);

-- pgcheck_nolint:missing-if-not-exists
CREATE TABLE pgcheck (id text PRIMARY KEY);

CREATE TABLE IF NOT EXISTS pgcheck (id text PRIMARY KEY);

CREATE INDEX CONCURRENTLY pgcheck_idx ON pgcheck(id);

-- pgcheck_nolint:missing-if-not-exists
CREATE INDEX CONCURRENTLY pgcheck_idx ON pgcheck(id);

-- index without name should succeed
CREATE INDEX CONCURRENTLY ON pgcheck(id);
