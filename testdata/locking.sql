CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

-- pgcheck_nolint:non-concurrent-index-creation
CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ON pgcheck(value);

ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

-- pgcheck_nolint:constraint-excessive-lock
ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;
