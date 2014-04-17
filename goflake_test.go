package goflake

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
		_, err := NewGoFlake(v)
		assert.Equal(t, err, nil, "Error should be nil")
	}
}

func TestInvalidWorkerId(t *testing.T) {
	invalidIds := []uint32{1024, 5841, 892347934}
	for _, v := range invalidIds {
		_, err := NewGoFlake(v)
		assert.Equal(t, err, ErrInvalidWorkerId, "Error should match")
	}
}

func TestGenerate(t *testing.T) {
	gf, err := NewGoFlake(0)
	assert.Equal(t, err, nil, "Error should be nil")

	id, err := gf.Generate()
	assert.Equal(t, err, nil, "Error should be nil")
}
