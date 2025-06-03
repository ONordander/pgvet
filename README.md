# pgvet 🛡️

pgvet is a database migration linter for [PostgreSQL](https://www.postgresql.org/).

Avoid problematic migrations and application downtime by detecting:

- Failing changes
- Non backwards compatible changes
- Migrations that use excessive locking or risk deadlocks
- Non idempotent changes

![Recording](docs/pgvet.gif)

Available as binary and as a [Github Action](#github-action)

# Installation

Prebuilt binaries for Linux, macOS, and Windows are available under [releases](https://github.com/onordander/pgvet/releases)

Installing with Golang is also possible:

```shell
CGO_ENABLED=0 go install github.com/onordander/pgvet@latest
```

## Github action

```yaml
steps:
- name: pgvet
  uses: onordander/pgvet@v0.2.1
  with:
    pattern: "./migrations/*.sql"
    config: "./pgvet.yaml"
```

# Usage

```sql
-- migrations/001.sql
ALTER TABLE pgvet ADD COLUMN name text NOT NULL;

CREATE INDEX pgvet_name_key ON pgvet(name);
```

```shell
⇥ pgvet lint migrations/*.sql

add-non-null-column: migrations/001.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgvet ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#add-non-null-column
........................................................................................................................

non-concurrent-index: migrations/001.sql:5

  5 | CREATE INDEX pgvet_name_key ON pgvet(name)

  Violation: Creating/dropping an index non-concurrently acquires a lock on the table that block writes for the duration of the operation
  Solution: Create/drop the index concurrently using the `CONCURRENTLY` option to avoid blocking. Note: this cannot be done inside a transaction
  Explanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#non-concurrent-index
........................................................................................................................

missing-if-not-exists: migrations/001.sql:5

  5 | CREATE INDEX pgvet_name_key ON pgvet(name)

  Violation: Creating an object might fail if it already exists, making the migration non idempotent
  Solution: Wrap the create statements with guards; e.g. CREATE TABLE IF NOT EXISTS pgvet ...
  Explanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#missing-if-not-exists
........................................................................................................................
```

## JSON formatting

```shell
⇥ pgvet lint --format=json migrations/001.sql

[{"file":"migrations/001.sql","code":"add-non-null-column","statement":"-- migrations/001.sql\nALTER TABLE pgvet ADD COLUMN name text NOT NULL","statementLine":1,"slug":"Adding a non-nullable column without a default will fail if the table is populated","help":"Make the column nullable or add a default"},{"file":"migration.sql","code":"non-concurrent-index","statement":"CREATE INDEX pgvet_name_key ON pgvet(name)","statementLine":4,"slug":"Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built","help":"Build the index concurrently to avoid blocking. Note: this cannot be done inside a transaction"},{"file":"migration.sql","code":"missing-if-not-exists","statement":"CREATE INDEX pgvet_name_key ON pgvet(name)","statementLine":4,"slug":"Creating an object might fail if it already exists, making the migration non idempotent","help":"Wrap the create statements with guards; e.g. CREATE TABLE IF NOT EXISTS pgvet ..."}]
```

## Disabling rules with configuration

```yaml
# config.yaml
rules:
  missing-if-not-exists:
    enabled: false
  non-concurrent-index:
    enabled: false
```

```shell
⇥  pgvet lint --config=config.yaml migrations/001.sql

add-non-null-column: migrations/*.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgvet ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#add-non-null-column
........................................................................................................................
```

## Disabling with nolint directives

```sql
-- migration.sql
ALTER TABLE pgvet ADD COLUMN name text NOT NULL;

-- pgvet_nolint:non-concurrent-index,missing-if-not-exists
CREATE INDEX pgvet_name_key ON pgvet(name);
```

```shell
⇥  pgvet lint migration.sql

add-non-null-column: migration.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgvet ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#add-non-null-column
........................................................................................................................
```

# Rules

For examples see `./testdata`.

| Rule                                                          | Category      | Enabled by default |
| --------------------------------------------------------------| --------------| ------------------ |
| [drop-column](#drop-column)                                   | breaking      | ✓                  |
| [drop-table](#drop-table)                                     | breaking      | ✓                  |
| [rename-column](#rename-column)                               | breaking      | ✓                  |
| [change-column-type](#change-column-type)                     | breaking      | ✓                  |
| [add-non-null-column](#add-non-null-column)                   | nullability   | ✓                  |
| [set-non-null-column](#set-non-null-column)                   | nullability   | ✓                  |
| [non-concurrent-index](#non-concurrent-index)                 | locking       | ✓                  |
| [constraint-excessive-lock](#constraint-excessive-lock)       | locking       | ✓                  |
| [multiple-locks](#multiple-locks)                             | locking       | 🗙                  |
| [missing-if-not-exists](#missing-if-not-exists)               | idempotency   | ✓                  |
| [missing-if-exists](#missing-if-exists)                       | idempotency   | ✓                  |
| [use-timestamp-with-time-zone](#use-timestamp-with-time-zone) | types         | ✓                  |
| [missing-foreign-key-index](#missing-foreign-key-index)       | miscellaneous | ✓                  |

## Breaking changes

### drop-column

Enabled by default: ✓

Dropping a column is not backwards compatible and may break existing clients that depend on the column.

**Violation:**

```sql
ALTER TABLE pgvet DROP COLUMN id;
```

**Solution**:

1. Update the application code to no longer use the column
1. Ignore the violation by adding a nolint directive: `-- pgvet_nolint:drop-column`

***

### drop-table

Enabled by default: ✓

Dropping a table is not backwards compatible and may break existing clients that depend on the table.

**Violation:**

```sql
DROP TABLE pgvet;
```

**Solution**:

1. Update the application code to no longer use the table
1. Ignore the violation by adding a nolint directive: `-- pgvet_nolint:drop-table`

***

### rename-column

Enabled by default: ✓

Renaming a column is not backwards compatible and may break existing clients that depend on the old column name.

**Violation:**

```sql
ALTER TABLE pgvet RENAME name TO reference;
```

**Solution**:

1. Create a new column with the new name
1. Update the application to write to both columns
1. Copy the data from the old column to the new column
1. Update the application to only use the new column
1. Drop the old column

***

### change-column-type

Enabled by default: ✓

Changing the type of a column is not backwards compatible and may break existing clients that still expect the old type.

**Solution**:

1. Create a new column with the new type
1. Update the application to write to both columns
1. Copy the data from the old column to the new column
1. Update the application to only use the new column
1. Drop the old column

***

## Invalid null changes

### add-non-null-column

Enabled by default: ✓

Adding a non-nullable column without a default will fail if the table is populated.

**Violation:**

```sql
ALTER TABLE pgcheck ADD COLUMN value NOT NULL;
```

**Solution**:

*Option 1*: make the column nullable:
```sql
ALTER TABLE pgvet ADD COLUMN value text;
```

*Option 2*: give the column a default:

```sql
ALTER TABLE pgvet ADD COLUMN value text DEFAULT '1';
```

***

### set-non-null-column

Enabled by default: ✓

Altering a column to be non-nullable might fail if the column contains null values.

**Violation:**

```sql
ALTER TABLE pgvet ALTER COLUMN value SET NOT NULL;
```

**Solution**:

1. Ensure that the application always inserts a value.
1. Ensure that the column contains no nulls:
    ```sql
    SELECT COUNT(1) FROM pgvet WHERE value IS NULL;
    ```
1. Ignore the violation by adding a nolint directive: `-- pgvet_nolint:set-non-null-column`

***

## Locking

### non-concurrent-index

Enabled by default: ✓

Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built.

See: [Postgres - explicit locking](https://www.postgresql.org/docs/current/explicit-locking.html) (ShareLock)

**Violation:**

```sql
CREATE INDEX IF NOT EXISTS pgvet_value_idx ON pgvet(value);
```

**Solution**:

Use the `CONCURRENTLY` option:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS pgvet_value_idx ON pgvet(value);
```

*Note*: this cannot be done inside a transaction.

***

### constraint-excessive-lock

Enabled by default: ✓

Adding a constraint acquires a lock blocking any writes (and potential reads) during the constraint validation.
Further, if the constraint is a foreign key reference it acquires a lock on both tables.

See [Postgres - add table constraint](https://www.postgresql.org/docs/current/sql-altertable.html#SQL-ALTERTABLE-DESC-ADD-TABLE-CONSTRAINT)

**Violation:**

```sql
ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);
```

**Solution**:

1. Add the constraint with the `NOT VALID` option forcing it to not validate the constraint initially. This is a very fast operation as no validation is needed.
    ```sql
    ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;
    ```
1. Validate the constraint in a subsequent transaction. This acquires a more relaxed lock that doesn't block reads or writes.
    ```sql
    ALTER TABLE pgvet VALIDATE CONSTRAINT reference_fk;
    ```

### multiple-locks

Enabled by default: 🗙

Experimental: acquiring multiple locks in a single transaction can cause a deadlock if an application contends with the locks in a different order.\
Note: this rule assumes that the migrations runs in an implicit transaction.

**Violation:**

```sql
-- migrations/001.sql
BEGIN;
ALTER TABLE pgvet ADD COLUMN value text; -- acquires an ACCESS EXCLUSIVE lock
ALTER TABLE othertable ADD COLUMN value text; -- tries to acquire an ACCESS EXCLUSIVE lock but has to wait for the application code to release its lock
COMMIT;
```

```sql
-- application code
BEGIN;
UPDATE othertable SET name = 'newname' WHERE id = 1; -- acquires a ROW EXCLUSIVE lock that conflicts with ACCESS EXCLUSIVE
UPDATE pgvet SET name = 'newname' WHERE id = 1; -- this fails because the migration has a lock on 'pgvet' and is waiting for a lock on 'othertable'
COMMIT;
```


See [Postgres - Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)

**Solution**:

Perform the changes in separate transactions.

```sql
-- migrations/001.sql
BEGIN;
ALTER TABLE pgvet ADD COLUMN value text;
COMMIT;

BEGIN;
ALTER TABLE othertable ADD COLUMN value text;
COMMIT;
```

## Idempotency

### missing-if-not-exists

Enabled by default: ✓

Creating an object might fail if it already exists, making the migration non idempotent.

**Violation:**

```sql
CREATE TABLE pgcheck (id text PRIMARY KEY);
```

**Solution**:

Use the `IF NOT EXISTS` option:

```sql
CREATE TABLE IF NOT EXISTS pgvet (id text PRIMARY KEY);
```

***

### missing-if-exists

Enabled by default: ✓

Dropping objects/relations might fail if they do not exist, making the migration non idempotent.

**Violation:**

```sql
DROP INDEX CONCURRENTLY pgvet_idx;
```

**Solution**:

Use the `IF EXISTS` option:

```sql
DROP INDEX CONCURRENTLY IF EXISTS pgvet_idx;
```

## Types

### use-timestamp-with-time-zone

Enabled by default: ✓

Timestamp with time zone preserves the time zone information and makes the data easier to reason about.

**Violation**:

```sql
CREATE TABLE IF NOT EXISTS pgvet (
  id text PRIMARY KEY,
  created_at timestamp
);
```

**Solution**:

Use `timestamptz`/`timestamp with time zone`

```sql
CREATE TABLE IF NOT EXISTS pgvet (
  id text PRIMARY KEY,
  created_at timestamptz
);
```

## Miscellaneous

### missing-foreign-key-index

Enabled by default: ✓

When adding a foreign key constraint PostgreSQL will not automatically create an index for you.\
The referenced column is often used in joins and lookups, and thus can benefit from an index.

**Violation:**

```sql
CREATE TABLE IF NOT EXISTS pgvet (
  id text PRIMARY KEY,
  reference text REFERENCES parent(id),
);
-- end of migration
```

**Solution**:

Create an index for the referenced column:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS pgvet_idx ON pgvet(reference);
```

# Further reading

- [PostgreSQL at Scale: Database Schema Changes Without Downtime](https://medium.com/paypal-tech/postgresql-at-scale-database-schema-changes-without-downtime-20d3749ed680)
- [PostgreSQL - Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [PostgreSQL - Alter Table](https://www.postgresql.org/docs/current/sql-altertable.html)
