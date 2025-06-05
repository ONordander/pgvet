COMMIT; -- Exit implicit transaction

--
-- rule: non-concurrent-index
--
CREATE INDEX IF NOT EXISTS pgvet_idx ON pgvet(value);

-- pgvet_nolint:non-concurrent-index
CREATE INDEX IF NOT EXISTS pgvet_idx ON pgvet(value);

CREATE INDEX CONCURRENTLY IF NOT EXISTS pgvet_idx ON pgvet(value);

DROP INDEX IF EXISTS pgvet_idx;
 
-- pgvet_nolint:non-concurrent-index
DROP INDEX IF EXISTS pgvet_idx;

DROP INDEX CONCURRENTLY IF EXISTS pgvet_idx;

--
-- rule: constraint-excessive-lock
--

BEGIN;

ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

-- pgvet_nolint:constraint-excessive-lock
ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);

ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;

COMMIT;

-- Add index to not violate missing-foreign-key-index
CREATE INDEX CONCURRENTLY ON pgvet(reference);


--
-- rule: multiple-locks
--

BEGIN;
ALTER TABLE firsttable ADD COLUMN IF NOT EXISTS value text;
ALTER TABLE secondtable ADD COLUMN IF NOT EXISTS value text;

-- pgvet_nolint:multiple-locks
ALTER TABLE thirdtable ADD COLUMN IF NOT EXISTS value text;

COMMIT;
ALTER TABLE fourthtable ADD COLUMN IF NOT EXISTS value text;
