package bigflake

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mattheath/kala/util"
)

var bigId *BigflakeId

func TestMintBigflakeId(t *testing.T) {
	var err error
	bf := newBigflakeMinter(t)

	var id *BigflakeId
	for i := 0; i < 10; i++ {
		id, err = bf.Mint()
		t.Log(id.String())
		assert.NoError(t, err)
	}
	bigId = id
}

func TestParseBigFlake(t *testing.T) {
	testCases := []struct {
		lastTs   int64
		workerId int64
		sequence int64
	}{
		{1397666977000, 0, 0},     // Plain bit shift 22
		{2344466898000, 0, 0},     // Plain bit shift 22
		{1397666977000, 10, 0},    // Worker 10
		{2344466898000, 10, 0},    // Worker 10
		{1397666977000, 1023, 0},  // Worker 1023
		{2344466898000, 1023, 0},  // Worker 1023
		{1397666977000, 10, 123},  // Worker 10 & Sequence 123
		{2344466898000, 10, 1230}, // Worker 10 & Sequence 1230
		{1397666977000, 10, 2356}, // Worker 10 & Sequence 2356
		{2344466898000, 10, 4090}, // Worker 10 & Sequence 4090
	}

	for _, tc := range testCases {
		id := MintId(tc.lastTs, tc.workerId, tc.sequence)
		ts, workerId, sequence := ParseId(id)

		assert.Equal(t, tc.lastTs, ts)
		assert.Equal(t, tc.workerId, workerId)
		assert.Equal(t, tc.sequence, sequence)
	}
}

func TestBigflakeSnowflakeMintCompatibility(t *testing.T) {
	testCases := []struct {
		lastTs   int64
		workerId int64
		sequence int64
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

	// Test that the bigflake minter generates snowflake compatible IDs
	// when provided with the same test cases
	for _, tc := range testCases {
		id := mintId(tc.lastTs, tc.workerId, tc.sequence, 10, 12)
		assert.Equal(t, uint64(tc.id), id.Uint64(), fmt.Sprintf("IDs should match. Provided: '%d', Returned: '%s' ", tc.id, id))
	}
}

func BenchmarkMintBigflakeId(b *testing.B) {
	var id *big.Int
	var lastTs, workerId, sequenceId int64

	// Setup
	lastTs, workerId, sequenceId = 1397666977000, 10, 2356

	// Zoom!
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		id = mintId(lastTs, workerId, sequenceId, 10, 12)
	}

	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	bigId = &BigflakeId{id: id}
}

func newBigflakeMinter(t *testing.T) *Bigflake {
	mac := "80:36:bc:db:64:16"
	workerId, err := util.MacAddressToWorkerId(mac)
	if err != nil {
		t.Fail()
	}

	return &Bigflake{
		lastTimestamp: util.TimeToMsInt64(time.Now()),

		workerIdBits: defaultWorkerIdBits,
		sequenceBits: defaultSequenceBits,

		workerId: int64(workerId),
		sequence: 0,
		epoch:    0,
	}
}
