CREATE TABLE pgvet (id text PRIMARY KEY);

-- pgvet_nolint:missing-if-not-exists
CREATE TABLE pgvet (id text PRIMARY KEY);

CREATE TABLE IF NOT EXISTS pgvet (id text PRIMARY KEY);

-- pgvet_nolint:non-concurrent-index
CREATE INDEX pgvet_idx ON pgvet(id);

-- pgvet_nolint:missing-if-not-exists
-- pgvet_nolint:non-concurrent-index
CREATE INDEX pgvet_idx ON pgvet(id);

-- index without name should succeed
-- pgvet_nolint:non-concurrent-index
CREATE INDEX ON pgvet(id);

ALTER TABLE pgvet ADD COLUMN value text;

-- pgvet_nolint:missing-if-not-exists
ALTER TABLE pgvet ADD COLUMN value text;

-- pgvet_nolint:drop-table
DROP TABLE pgvet;
-- pgvet_nolint:drop-table
DROP TABLE IF EXISTS pgvet;

-- pgvet_nolint:non-concurrent-index
DROP INDEX pgvet_idx;
-- pgvet_nolint:non-concurrent-index
DROP INDEX IF EXISTS pgvet_idx;

-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN id;
-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN IF EXISTS id;
