COMMIT; -- exit implicit transaction

CREATE INDEX CONCURRENTLY pgvet_idx ON pgvet(id);

ALTER TABLE pgvet ADD COLUMN value text NOT NULL;
