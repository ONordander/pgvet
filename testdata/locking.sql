BEGIN;

--
-- rule: non-concurrent-index
--
CREATE INDEX IF NOT EXISTS pgvet_idx ON pgvet(value);

-- pgvet_nolint:non-concurrent-index
CREATE INDEX IF NOT EXISTS pgvet_idx ON pgvet(value);

CREATE INDEX CONCURRENTLY IF NOT EXISTS pgvet_idx ON pgvet(value);

DROP INDEX pgvet_idx;
 
-- pgvet_nolint:non-concurrent-index
DROP INDEX pgvet_idx;

DROP INDEX CONCURRENTLY pgvet_idx;

COMMIT;

--
-- rule: constraint-excessive-lock
--

BEGIN;

ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

-- pgvet_nolint:constraint-excessive-lock
ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;

-- Add index to not violate missing-foreign-key-index
CREATE INDEX CONCURRENTLY ON pgvet(reference);

COMMIT;

--
-- rule: multiple-locks
--

BEGIN;
ALTER TABLE firsttable ADD COLUMN value text;
ALTER TABLE secondtable ADD COLUMN value text;

-- pgvet_nolint:multiple-locks
ALTER TABLE thirdtable ADD COLUMN value text;

COMMIT;
ALTER TABLE fourthtable ADD COLUMN value text;
