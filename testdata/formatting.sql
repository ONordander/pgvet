


-- This is a comment



ALTER TABLE pgvet DROP COLUMN value;
-- pgvet_nolint:rename-column
ALTER TABLE pgvet
  RENAME COLUMN value
  TO newvalue;

ALTER TABLE pgvet
  RENAME COLUMN
  value
  TO
  newvalue;
-- This is another comment
