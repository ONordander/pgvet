ALTER TABLE pgvet DROP COLUMN IF EXISTS value;

-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN IF EXISTS nolint;

ALTER TABLE pgvet RENAME column oldvalue TO newvalue;

-- pgvet_nolint:rename-column safe: has been removed
ALTER TABLE pgvet RENAME column oldvalue TO newvalue;

DROP TABLE IF EXISTS pgvet;

-- pgvet_nolint:drop-table
DROP TABLE IF EXISTS pgvet;

ALTER TABLE pgvet RENAME TO pgvet_new;

-- pgvet_nolint:rename-table
ALTER TABLE pgvet RENAME TO pgvet_new;

ALTER TABLE pgvet ALTER COLUMN value TYPE text;

-- pgvet_nolint:change-column-type
ALTER TABLE pgvet ALTER COLUMN value TYPE text;
