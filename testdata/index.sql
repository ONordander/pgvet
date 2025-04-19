CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

-- pgcheck_nolint:non-concurrent-index-creation
CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ON pgcheck(value);
