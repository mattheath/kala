package kala

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

	for _, tc := range testCases {
		adjTs := customTimestamp(time.Unix(tc.ts/1000, 0))
		assert.Equal(t, adjTs, tc.adjTs, "Times should match")
	}
}

func TestValidWorkerId(t *testing.T) {
	validIds := []uint32{0, 545, 1023}
	for _, v := range validIds {
		_, err := NewSnowflake(v)
		assert.Equal(t, err, nil, "Error should be nil")
	}
}

func TestInvalidWorkerId(t *testing.T) {
	invalidIds := []uint32{1024, 5841, 892347934}
	for _, v := range invalidIds {
		_, err := NewSnowflake(v)
		assert.Equal(t, err, ErrInvalidWorkerId, "Error should match")
	}
}

func TestSequenceOverflow(t *testing.T) {
	invalidSequenceIds := []uint32{4096, 5841, 892347934}
	for _, seq := range invalidSequenceIds {
		gf := &goFlake{
			lastTimestamp: customTimestamp(time.Now()), // YUK
			workerId:      0,
			sequence:      seq,
		}
		_, err := gf.Mint()
		assert.Equal(t, err, ErrSequenceOverflow, "Error should match")
	}
}

func TestMint(t *testing.T) {
	gf, err := NewSnowflake(0)
	assert.Equal(t, err, nil, "Error should be nil")

	_, err = gf.Mint()
	assert.Equal(t, err, nil, "Error should be nil")
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
		gf := &goFlake{
			lastTimestamp: tc.lastTs,
			workerId:      tc.workerId,
			sequence:      tc.sequence,
		}
		id := gf.mintId()
		assert.Equal(t, id, tc.id, fmt.Sprintf("IDs should match. Provided: '%s', Returned: '%s' ", tc.id, id))
	}
}

func TestBackwardsTimeError(t *testing.T) {
	gf := &goFlake{lastTimestamp: 1397666977000}
	err := gf.update(1397666976999)
	assert.NotEqual(t, err, nil, "Error should not be nil")
}

func TestTimeOverflow(t *testing.T) {
	gf := &goFlake{lastTimestamp: 1397666977000}
	err := gf.update(2199023255552)
	assert.Equal(t, err, ErrOverflow, "Errors should match")
}

func TestPreEpochTime(t *testing.T) {
	testCases := []time.Time{
		time.Date(2012, 1, 0, 0, 0, 0, 0, time.UTC),
		time.Date(2011, 9, 5, 0, 0, 0, 0, time.UTC),
		time.Date(1066, 9, 5, 0, 0, 0, 0, time.UTC),
	}
	for _, tc := range testCases {
		gf := &goFlake{}
		ts := customTimestamp(tc)
		err := gf.update(ts)
		assert.NotEqual(t, err, nil, "Errors should match")
	}
}
