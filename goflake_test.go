package goflake

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCustomTimestamp(t *testing.T) {

	// 2014-04-16 17:49:37 +0100 => 1397666977
	tm := time.Unix(1397666977, 0)
	newT := customTimestamp(tm)

	// timestamp - epoch = adjusted time
	// 1397666977000 - 1325376000000 = 72290977000
	assert.Equal(t, newT, 72290977000, "Times should match")
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
