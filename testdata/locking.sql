BEGIN;

--
-- rule: non-concurrent-index-creation
--
CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

-- pgcheck_nolint:non-concurrent-index-creation
CREATE INDEX IF NOT EXISTS pgcheck_idx ON pgcheck(value);

CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ON pgcheck(value);

COMMIT;

--
-- rule: constraint-excessive-lock
--

BEGIN;

ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

-- pgcheck_nolint:constraint-excessive-lock
ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;

-- Add index to not violate missing-foreign-key-index
CREATE INDEX CONCURRENTLY ON pgcheck(reference);

COMMIT;

--
-- rule: many-alter-table
--

BEGIN;
ALTER TABLE firsttable ADD COLUMN value text;
ALTER TABLE secondtable ADD COLUMN value text;

-- pgcheck_nolint:many-alter-table
ALTER TABLE thirdtable ADD COLUMN value text;

COMMIT;
ALTER TABLE fourthtable ADD COLUMN value text;
