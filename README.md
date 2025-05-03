# pgcheck

pgcheck is a database migration linter for [PostgreSQL](https://www.postgresql.org/).

It aims to aid in writing migrations according to best practices and detect:

- Failing changes
- Non backwards compatible changes
- Migrations that use excessive locking or risk deadlocks
- Non idempotent changes

# Installation

Prebuilt binaries are available under [releases](https://github.com/onordander/pgcheck/releases)

Installing with Golang is also possible:

```shell
CGO_ENABLED=0 go install github.com/onordander/pgcheck@latest
```

# Usage

```sql
-- migrations/001.sql
ALTER TABLE pgcheck ADD COLUMN name text NOT NULL;

CREATE INDEX pgcheck_name_key ON pgcheck(name);
```

```shell
⇥ pgcheck lint migrations/*.sql

add-non-null-column: migrations/001.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgcheck ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#add-non-null-column
........................................................................................................................

non-concurrent-index-creation: migrations/001.sql:5

  5 | CREATE INDEX pgcheck_name_key ON pgcheck(name)

  Violation: Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built.
  Solution: Build the index concurrently to avoid blocking. Note: this cannot be done inside a transaction
  Explanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#non-concurrent-index-creation
........................................................................................................................

missing-index-if-not-exists: migrations/001.sql:5

  5 | CREATE INDEX pgcheck_name_key ON pgcheck(name)

  Violation: Creating a named index will fail if it already exists, making the migration non idempotent
  Solution: Wrap the create statements with guards; e.g. CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ...
  Explanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#missing-index-if-not-exists
........................................................................................................................
```

## JSON formatting

```shell
⇥ pgcheck lint --format=json migrations/001.sql

[{"file":"migrations/001.sql","code":"add-non-null-column","statement":"-- migrations/001.sql\nALTER TABLE pgcheck ADD COLUMN name text NOT NULL","statementLine":1,"slug":"Adding a non-nullable column without a default will fail if the table is populated","help":"Make the column nullable or add a default"},{"file":"migration.sql","code":"non-concurrent-index-creation","statement":"CREATE INDEX pgcheck_name_key ON pgcheck(name)","statementLine":4,"slug":"Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built","help":"Build the index concurrently to avoid blocking. Note: this cannot be done inside a transaction"},{"file":"migration.sql","code":"missing-index-if-not-exists","statement":"CREATE INDEX pgcheck_name_key ON pgcheck(name)","statementLine":4,"slug":"Creating a named index will fail if it already exists, making the migration non idempotent","help":"Wrap the create statements with guards; e.g. CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ..."}]
```

## Disabling rules with configuration

```yaml
# config.yaml
rules:
  missing-index-if-not-exists:
    enabled: false
  non-concurrent-index-creation:
    enabled: false
```

```shell
⇥  pgcheck lint --config=config.yaml migrations/001.sql

add-non-null-column: migrations/*.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgcheck ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#add-non-null-column
........................................................................................................................
```

## Disabling with nolint directives

```sql
-- migration.sql
ALTER TABLE pgcheck ADD COLUMN name text NOT NULL;

-- pgcheck_nolint:non-concurrent-index-creation,missing-index-if-not-exists
CREATE INDEX pgcheck_name_key ON pgcheck(name);
```

```shell
⇥  pgcheck lint migration.sql

add-non-null-column: migration.sql:1

  1 | -- migrations/001.sql
  2 | ALTER TABLE pgcheck ADD COLUMN name text NOT NULL

  Violation: Adding a non-nullable column without a default will fail if the table is populated
  Solution: Make the column nullable or add a default
  Explanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#add-non-null-column
........................................................................................................................
```

# Rules

For examples see `./testdata`.

## Breaking changes

### drop-column

Enabled by default: ✓

Dropping a column is not backwards compatible and may break existing clients that depend on the column.

**Solution**:

1. Update the application code to no longer use the column
1. Ignore the violation by adding a nolint directive: `-- pgcheck_nolint:drop-column`

***

### drop-table

Enabled by default: ✓

Dropping a table is not backwards compatible and may break existing clients that depend on the table.

**Solution**:

1. Update the application code to no longer use the table
1. Ignore the violation by adding a nolint directive: `-- pgcheck_nolint:drop-table`

***

### rename-column

Enabled by default: ✓

Renaming a column is not backwards compatible and may break existing clients that depend on the old column name.

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

**Solution**:

*Option 1*: make the column nullable:
```sql
ALTER TABLE pgcheck ADD COLUMN value text;
```

*Option 2*: give the column a default:

```sql
ALTER TABLE pgcheck ADD COLUMN value text DEFAULT '1';
```

***

### set-non-null-column

Enabled by default: ✓

Altering a column to be non-nullable might fail if the column contains null values.

**Solution**:

1. Ensure that the column contains no nulls:
    ```sql
    SELECT COUNT(1) FROM pgcheck WHERE value IS NULL;
    ```
1. Ignore the violation by adding a nolint directive: `-- pgcheck_nolint:set-non-null-column`

***

## Locking

### non-concurrent-index-creation

Enabled by default: ✓

Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built.

See: [Postgres - explicit locking](https://www.postgresql.org/docs/current/explicit-locking.html) (ShareLock)

**Solution**:

Use the `CONCURRENTLY` option:

```sql
CREATE INDEX CONCURRENTLY pgcheck_value_idx ON pgcheck(value);
```

*Note*: this cannot be done inside a transaction.

***

### constraint-excessive-lock

Enabled by default: ✓

Adding a constraint acquires a lock blocking any writes (and potential reads) during the constraint validation.
Further, if the constraint is a foreign key reference it acquires a lock on both tables.

See [Postgres - add table constraint](https://www.postgresql.org/docs/current/sql-altertable.html#SQL-ALTERTABLE-DESC-ADD-TABLE-CONSTRAINT)

**Solution**:

1. Add the constraint with the `NOT VALID` option forcing it to not validate the constraint initially. This is a very fast operation as no validation is needed.
    ```sql
    ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;
    ```
1. Validate the constraint in a subsequent transaction. This acquires a more relaxed lock that that does not block reads or writes.
    ```sql
    ALTER TABLE pgcheck VALIDATE CONSTRAINT reference_fk;
    ```

## Idempotency

### missing-relation-if-not-exists

Enabled by default: ✓

Creating a relation will fail if it already exists, making the migration non idempotent.

**Solution**:

Use the `IF NOT EXISTS` option:

```sql
CREATE TABLE IF NOT EXISTS pgcheck (id text PRIMARY KEY);
```

***

### missing-index-if-not-exists

Enabled by default: ✓

Creating a named index will fail if it already exists, making the migration non idempotent.

**Solution**:

Use the `IF NOT EXISTS` option:

```sql
CREATE INDEX IF NOT EXISTS pgcheck_value_idx ON pgcheck(value);
```

***

## Miscellaneous

### missing-foreign-key-index

Enabled by default: 🗙

When adding a foreign key constraint PostgreSQL will not automatically create an index for you.\
The referenced column is often used in joins and lookups, and thus can benefit from an index.

**Solution**:

Create an index for the referenced column:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ON pgcheck(reference);
```
