CREATE TABLE IF NOT EXISTS pgvet (
  id text PRIMARY KEY,
  reference text REFERENCES parent(id),
  other_reference text REFERENCES parent(id)
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS ref_fk ON pgvet(reference);

-- pgvet_nolint:missing-foreign-key-index
CREATE TABLE IF NOT EXISTS pgvet_two (
  id text PRIMARY KEY,
  reference text REFERENCES parent(id),
  other_reference text REFERENCES parent(id)
);
