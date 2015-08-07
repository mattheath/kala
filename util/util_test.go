package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMacAddressToWorkerId(t *testing.T) {
	mac := "80:36:bc:db:64:16"
	workerId, err := MacAddressToWorkerId(mac)
	require.NoError(t, err)
	assert.EqualValues(t, 140972585083926, workerId)
}

func TestCustomTimestamp(t *testing.T) {
	// Formatted using the default format string
	// "2006-01-02 15:04:05.999999999 -0700 MST"
	testcases := []struct {
		epoch     int64
		timestamp string
		expected  int64
	}{
		{0, "2012-01-01 00:00:00.000000000 +0000 UTC", 1325376000000},
		{1325376000000, "2012-01-01 00:00:00.000000000 +0000 UTC", 0},
		{1325376000000, "2012-01-01 00:00:00.000000001 +0000 UTC", 0},
		{1325376000000, "2012-01-01 00:00:00.001000001 +0000 UTC", 1},
		{1325376000000, "2011-12-31 23:59:59.999999999 +0000 UTC", -1},          // will be floored to .999
		{1325376000000, "2013-01-01 00:00:00.000000000 +0000 UTC", 31622400000}, // 2012 was a leap year
		{1325376000000, "2013-01-01 00:00:00.000000000 +0000 UTC", 31622400000},
		{1325376000000, "2081-09-06 15:47:35 +0000 UTC", 2199023255000},
		{1325376000000, "2840-04-02 20:16:16.530845939 +0000 UTC", 26137196176530},
	}

	for _, tc := range testcases {
		ts, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tc.timestamp)
		require.NoError(t, err)
		ct := CustomTimestamp(tc.epoch, ts)
		t.Logf("%v\t%v\t%v", tc.epoch, ts, ct)
		assert.Equal(t, tc.expected, ct)
	}
}

func TestTimeToMsInt64(t *testing.T) {
	// Formatted using the default format string
	// "2006-01-02 15:04:05.999999999 -0700 MST"
	testcases := []struct {
		timestamp string
		expected  int64
	}{
		{"2015-04-02 20:16:16.000000000 +0000 UTC", 1428005776000},
		{"2015-04-02 20:16:16.530845939 +0000 UTC", 1428005776530},
		{"2011-03-01 12:45:01.839 +0000 UTC", 1298983501839},
		{"2011-03-01 12:45:01.8391 +0000 UTC", 1298983501839},
		{"1970-01-01 00:00:00.000 +0000 UTC", 0},
		{"1969-12-31 23:59:59.999 +0000 UTC", -1},
		{"1700-04-03 17:59:54.000000000 +0000 UTC", -8512322406000},

		// Maximum time which can be stored by signed int64 in nanoseconds
		{"2262-04-11 23:47:16.854775807 +0000 UTC", 9223372036854},
		// One nanosecond higher causes overflow with previously flawed implementation
		{"2262-04-11 23:47:16.854775808 +0000 UTC", 9223372036854},

		// Minimum time which can be stored by signed int64 in nanoseconds
		{"1677-09-21 00:12:43.145224192 +0000 UTC", -9223372036855},
		// One nanosecond lower overflows int64 with previously flawed implementation
		{"1677-09-21 00:12:43.145224191 +0000 UTC", -9223372036855},
	}

	for _, tc := range testcases {
		ts, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tc.timestamp)
		require.NoError(t, err)

		ms := TimeToMsInt64(ts)
		t.Logf("%s :: %v", tc.timestamp, ms)

		assert.Equal(t, tc.expected, ms, fmt.Sprintf("Expected %s", tc.timestamp))

		ts2 := MsInt64ToTime(ms)
		assert.Equal(t, ts.Truncate(time.Millisecond).String(), ts2.Truncate(time.Millisecond).String())
	}
}
