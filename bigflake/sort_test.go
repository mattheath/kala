package bigflake

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/mattheath/base62"
	"github.com/mattheath/kala/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBase62KSortability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping base62 k-ordering tests")
	}

	// Use a custom encoder so we can set the padding
	e := base62.NewStdEncoding().Option(base62.Padding(25))

	ksortability(t, func(id *BigflakeId) string {
		return e.EncodeBigInt(id.Raw())
	})
}

func TestStringKSortability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping uuid k-ordering tests")
	}

	ksortability(t, func(id *BigflakeId) string {
		return id.String()
	})
}

func TestUuidKSortability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping uuid k-ordering tests")
	}

	ksortability(t, func(id *BigflakeId) string {
		return id.Uuid()
	})
}

func ksortability(t *testing.T, formatFunc func(id *BigflakeId) string) {
	var (
		lexicalOrder  sort.StringSlice = make([]string, 0)
		originalOrder                  = make([]string, 0)
		id                             = &BigflakeId{}

		// Allow us to progressively jump forwards in time
		timeDiff time.Duration = 10 * time.Millisecond
	)

	// Generate lots of ids
	bf, err := New(0)
	require.NoError(t, err)
	bf.setup()

	for i := 0; i < 10000000; i++ {
		if i%300000 == 0 {
			timeDiff = timeDiff * 2
			t.Logf("Moved to %v offset", timeDiff)
		}

		// Update time, sequence etc
		err := bf.update(util.TimeToMsInt64(time.Now().Add(timeDiff)))
		require.NoError(t, err)
		id.id = MintId(bf.lastTimestamp, bf.workerId, bf.sequence)

		idStr := formatFunc(id)

		lexicalOrder = append(lexicalOrder, idStr)
		originalOrder = append(originalOrder, idStr)
	}

	// Sort string array
	lexicalOrder.Sort()

	// Compare ordering
	var mismatch int64
	for i, v := range originalOrder {
		if lexicalOrder[i] != v {
			mismatch++
		}
	}
	assert.Equal(t, int64(0), mismatch, fmt.Sprintf("Expected zero mismatches, got %v", mismatch))
}
