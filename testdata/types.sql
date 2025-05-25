CREATE TABLE IF NOT EXISTS pgvet (
  created_at timestamp
);

ALTER TABLE pgvet ADD COLUMN IF NOT EXISTS updated_at timestamp;

CREATE TABLE IF NOT EXISTS pgvet (
  created_at timestamptz
);

ALTER TABLE pgvet ADD COLUMN IF NOT EXISTS updated_at timestamptz;
