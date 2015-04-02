package kala

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var bigId *big.Int

func TestMintBigflakeId(t *testing.T) {

	mac := "80:36:bc:db:64:16"
	workerId, err := MacAddressToWorkerId(mac)
	if err != nil {
		t.Fail()
	}

	bf := &Bigflake{
		lastTimestamp: timeToMsInt64(time.Now()),

		workerIdBits: defaultBigflakeWorkerIdBits,
		sequenceBits: defaultBigflakeSequenceBits,

		workerId: int64(workerId),
		sequence: 0,
		epoch:    0,
	}

	start := time.Now()

	var id *big.Int
	for i := 0; i < 10; i++ {
		id, err = bf.Mint()
		t.Log(id.String())
		assert.NoError(t, err)
	}

	bigId = id

	fmt.Println("elapsed: ", time.Since(start))

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
		bf := &Bigflake{}
		id := bf.mintId(tc.lastTs, tc.workerId, tc.sequence, 10, 12)
		assert.Equal(t, uint64(tc.id), id.Uint64(), fmt.Sprintf("IDs should match. Provided: '%s', Returned: '%s' ", tc.id, id))
	}
}

func BenchmarkMintBigflakeId(b *testing.B) {
	var id *big.Int
	var lastTs, workerId, sequenceId int64

	// Setup
	lastTs, workerId, sequenceId = 1397666977000, 10, 2356
	bf := &Bigflake{}

	// Zoom!
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		id = bf.mintId(lastTs, workerId, sequenceId, 10, 12)
	}

	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	bigId = id
}
