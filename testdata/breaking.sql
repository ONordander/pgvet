ALTER TABLE pgcheck DROP COLUMN value;

-- pgcheck_nolint:drop-column
ALTER TABLE pgcheck DROP COLUMN nolint;

ALTER TABLE pgcheck RENAME column oldvalue TO newvalue;

-- pgcheck_nolint:rename-column safe: has been removed
ALTER TABLE pgcheck RENAME column oldvalue TO newvalue;

DROP TABLE pgcheck;

-- pgcheck_nolint:drop-table
DROP TABLE pgcheck;

ALTER TABLE pgcheck ALTER COLUMN value TYPE text;

-- pgcheck_nolint:change-column-type
ALTER TABLE pgcheck ALTER COLUMN value TYPE text;
