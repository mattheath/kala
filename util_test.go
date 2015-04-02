package kala

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMacAddressToWorkerId(t *testing.T) {
	mac := "80:36:bc:db:64:16"
	workerId, err := MacAddressToWorkerId(mac)
	require.NoError(t, err)
	assert.Equal(t, uint(140972585083926), workerId)
}
