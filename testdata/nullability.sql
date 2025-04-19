ALTER TABLE pgcheck ADD COLUMN value text NOT NULL;

-- pgcheck_nolint:add-non-null-column
ALTER TABLE pgcheck ADD COLUMN value text NOT NULL;

ALTER TABLE pgcheck ALTER COLUMN nullvalue SET NOT NULL;

-- pgcheck_nolint:set-non-null-column
ALTER TABLE pgcheck ALTER COLUMN nullvalue SET NOT NULL;

ALTER TABLE pgcheck
  ALTER COLUMN nullvalue SET NOT NULL,
  ADD COLUMN nonnull text NOT NULL;

-- pgcheck_nolint:set-non-null-column,add-non-null-column
ALTER TABLE pgcheck
  ALTER COLUMN nullvalue SET NOT NULL,
  ADD COLUMN nonnull text NOT NULL;
