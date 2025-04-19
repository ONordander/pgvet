


-- This is a comment



ALTER TABLE pgcheck DROP COLUMN value;
-- pgcheck_nolint:rename-column
ALTER TABLE pgcheck
  RENAME COLUMN value
  TO newvalue;

ALTER TABLE pgcheck
  RENAME COLUMN
  value
  TO
  newvalue;
-- This is another comment
