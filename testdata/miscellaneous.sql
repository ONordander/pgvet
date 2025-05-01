CREATE TABLE IF NOT EXISTS pgcheck (
  id text PRIMARY KEY,
  reference text REFERENCES parent(id),
  other_reference text REFERENCES parent(id)
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS ref_fk ON pgcheck(reference);

-- pgcheck_nolint:missing-foreign-key-index
CREATE TABLE IF NOT EXISTS pgcheck_two (
  id text PRIMARY KEY,
  reference text REFERENCES parent(id),
  other_reference text REFERENCES parent(id)
);
