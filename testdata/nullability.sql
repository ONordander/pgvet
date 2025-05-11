ALTER TABLE pgvet ADD COLUMN IF NOT EXISTS value text NOT NULL;

-- pgvet_nolint:add-non-null-column
ALTER TABLE pgvet ADD COLUMN IF NOT EXISTS value text NOT NULL;

ALTER TABLE pgvet ALTER COLUMN nullvalue SET NOT NULL;

-- pgvet_nolint:set-non-null-column
ALTER TABLE pgvet ALTER COLUMN nullvalue SET NOT NULL;

ALTER TABLE pgvet
  ALTER COLUMN nullvalue SET NOT NULL,
  ADD COLUMN IF NOT EXISTS nonnull text NOT NULL;

-- pgvet_nolint:set-non-null-column,add-non-null-column
ALTER TABLE pgvet
  ALTER COLUMN nullvalue SET NOT NULL,
  ADD COLUMN IF NOT EXISTS nonnull text NOT NULL;
