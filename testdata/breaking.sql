ALTER TABLE pgvet DROP COLUMN value;

-- pgvet_nolint:drop-column
ALTER TABLE pgvet DROP COLUMN nolint;

ALTER TABLE pgvet RENAME column oldvalue TO newvalue;

-- pgvet_nolint:rename-column safe: has been removed
ALTER TABLE pgvet RENAME column oldvalue TO newvalue;

DROP TABLE pgvet;

-- pgvet_nolint:drop-table
DROP TABLE pgvet;

ALTER TABLE pgvet ALTER COLUMN value TYPE text;

-- pgvet_nolint:change-column-type
ALTER TABLE pgvet ALTER COLUMN value TYPE text;
