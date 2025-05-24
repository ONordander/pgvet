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

-- pgvet_nolint:drop-table
DROP TABLE pgvet;
-- pgvet_nolint:drop-table
DROP TABLE IF EXISTS pgvet;

DROP INDEX CONCURRENTLY pgvet_idx;
DROP INDEX CONCURRENTLY IF EXISTS pgvet_idx;

-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN id;
-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN IF EXISTS id;
