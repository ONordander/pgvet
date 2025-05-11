package main

import (
	"testing"

	"github.com/onordander/pgvet/rules"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterNoLints(t *testing.T) {
	t.Parallel()

	t.Run("Should not filter mismatching rules", func(t *testing.T) {
		t.Parallel()
		q := `-- pgvet_nolint:a-rule
ALTER TABLE pgvet ADD COLUMN value text NOT NULL;`
		results := []rules.Result{
			{
				StmtStart: 0,
				StmtEnd:   72,
				Code:      "another-rule",
			},
		}

		assert.Len(t, filterNoLints(q, results), len(results))
	})

	t.Run("Should filter many violations", func(t *testing.T) {
		t.Parallel()
		q := `
-- pgvet_nolint:filter-this
CREATE INDEX terre ON terre (id);

-- pgvet_nolint:dont-filter-this
CREATE INDEX terreluring ON terre (id);

-- pgvet_nolint:filter-this-too,but-not-this
ALTER TABLE pgvet ADD COLUMN value text NOT NULL,
ALTER COLUMN value2 SET NOT NULL;`

		results := []rules.Result{
			{
				StmtStart: 0,
				StmtEnd:   60,
				Code:      "filter-this",
			},
			{
				StmtStart: 61,
				StmtEnd:   134,
				Code:      "wont-filter",
			},
			{
				StmtStart: 135,
				StmtEnd:   266,
				Code:      "filter-this-too",
			},
			{
				StmtStart: 135,
				StmtEnd:   266,
				Code:      "wont-filter-two",
			},
		}

		filtered := filterNoLints(q, results)
		require.Len(t, filtered, 2)
		assert.EqualValues(t, "wont-filter", filtered[0].Code)
		assert.EqualValues(t, 61, filtered[0].StmtStart)
		assert.EqualValues(t, "wont-filter-two", filtered[1].Code)
		assert.EqualValues(t, 135, filtered[1].StmtStart)
	})
}
