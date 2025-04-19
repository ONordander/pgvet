CREATE INDEX CONCURRENTLY pgcheck_idx ON pgcheck(id);

ALTER TABLE pgcheck ADD COLUMN value text NOT NULL;
