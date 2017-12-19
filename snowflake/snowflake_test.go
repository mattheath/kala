package snowflake

import (
	"fmt"
	"testing"
	"time"

	"github.com/mattheath/kala/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var result string

func TestCustomTimestamp(t *testing.T) {

	// timestamp - epoch = adjusted time
	testCases := []struct {
		ts    int64
		adjTs int64
	}{
		{1397666977000, 72290977000},   // Now
		{1397666978000, 72290978000},   // in 1 second
		{1395881056000, 70505056000},   // 3 weeks ago
		{1303001162000, -22374838000},  // 3 years ago
		{1492390054000, 167014054000},  // in 3 years
		{2344466898000, 1019090898000}, // in 30 years
	}

	// Initialise our custom epoch
	epoch, err := time.Parse(time.RFC3339, defaultEpoch)
	require.NoError(t, err)
	epochMs := util.TimeToMsInt64(epoch)

	for _, tc := range testCases {
		adjTs := util.CustomTimestamp(epochMs, time.Unix(tc.ts/1000, 0))
		assert.Equal(t, adjTs, tc.adjTs, "Times should match")
	}
}

func TestValidWorkerId(t *testing.T) {
	validWorkerIds := []uint32{0, 545, 1023}
	for _, v := range validWorkerIds {
		sf, err := New(v)
		require.NoError(t, err)

		id, err := sf.Mint()
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
	}
}

func TestInvalidWorkerId(t *testing.T) {
	invalidWorkerIds := []uint32{1024, 5841, 892347934}
	for _, v := range invalidWorkerIds {
		sf, err := New(v)
		require.NoError(t, err)

		id, err := sf.Mint()
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidWorkerId, "Error should match")
		assert.Equal(t, id, "")
	}
}

func TestSequenceOverflow(t *testing.T) {

	// Setup snowflake at a particular time which we will freeze at
	sf, err := New(0)
	require.NoError(t, err)
	tms := util.TimeToMsInt64(time.Now())
	sf.lastTimestamp = tms

	invalidSequenceIds := []uint32{4096, 5841, 892347934}
	for _, seq := range invalidSequenceIds {

		// Fix the sequence ID, then update
		// This should fail, as we are within the same ms
		sf.sequence = seq
		err := sf.update(tms)
		assert.Error(t, err)
		assert.Equal(t, err, ErrSequenceOverflow, "Error should match")
	}
}

func TestMint(t *testing.T) {
	sf, err := New(0)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		id, err := sf.Mint()
		t.Log(id)
		assert.NoError(t, err)
	}
}

func BenchmarkMintSnowflakeId(b *testing.B) {
	var id string

	sf, err := New(0)
	if err != nil {
		b.Fail()
	}

	// Ensure we can mint an id
	id, err = sf.Mint()
	if err != nil {
		b.Fail()
	}

	// Zoom!
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		id, _ = sf.Mint()
	}

	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	result = id
}

func TestMintId(t *testing.T) {
	testCases := []struct {
		lastTs   int64
		workerId uint32
		sequence uint32
		id       uint64
	}{
		{1397666977000, 0, 0, 5862240192299008000},     // Plain bit shift 22
		{2344466898000, 0, 0, 9833406888148992000},     // Plain bit shift 22
		{1397666977000, 10, 0, 5862240192299048960},    // Worker 10
		{2344466898000, 10, 0, 9833406888149032960},    // Worker 10
		{1397666977000, 1023, 0, 5862240192303198208},  // Worker 1023
		{2344466898000, 1023, 0, 9833406888153182208},  // Worker 1023
		{1397666977000, 10, 123, 5862240192299049083},  // Worker 10 & Sequence 123
		{2344466898000, 10, 1230, 9833406888149034190}, // Worker 10 & Sequence 1230
		{1397666977000, 10, 2356, 5862240192299051316}, // Worker 10 & Sequence 2356
		{2344466898000, 10, 4090, 9833406888149037050}, // Worker 10 & Sequence 4090
	}

	for _, tc := range testCases {
		sf, err := New(tc.workerId)
		require.NoError(t, err)

		sf.lastTimestamp = tc.lastTs
		sf.sequence = tc.sequence

		id := sf.mintId()
		assert.Equal(t, tc.id, id, fmt.Sprintf("IDs should match. Provided: '%v', Returned: '%v' ", tc.id, id))
	}
}

func TestBackwardsTimeError(t *testing.T) {
	sf, err := New(0)
	require.NoError(t, err)

	err = sf.update(1397666976999)
	assert.Error(t, err)
}

func TestTimeOverflow(t *testing.T) {
	sf, err := New(0)
	require.NoError(t, err)
	sf.lastTimestamp = 1397666977000

	err = sf.update(2199023255552)
	assert.Error(t, err)
	assert.Equal(t, err, ErrOverflow, "Errors should match")
}

func TestPreEpochTime(t *testing.T) {
	testCases := []time.Time{
		time.Date(2012, 1, 0, 0, 0, 0, 0, time.UTC),
		time.Date(2011, 9, 5, 0, 0, 0, 0, time.UTC),
		time.Date(1066, 9, 5, 0, 0, 0, 0, time.UTC),
	}
	for _, tc := range testCases {
		sf, err := New(0)
		require.NoError(t, err)

		// Initialise our custom epoch
		epoch, err := time.Parse(time.RFC3339, defaultEpoch)
		require.NoError(t, err)
		epochMs := util.TimeToMsInt64(epoch)
		ts := util.CustomTimestamp(epochMs, tc)

		err = sf.update(ts)
		assert.Error(t, err)
	}
}
