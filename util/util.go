package util

import (
	"math/big"
	"net"
	"time"
)

func MacAddressToWorkerId(mac string) (uint64, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return 0, err
	}

	workerId := new(big.Int).SetBytes([]byte(hw)).Uint64()

	return workerId, nil
}

// CustomTimestamp takes a timestamp and adjusts it to our custom epoch
func CustomTimestamp(epoch int64, t time.Time) int64 {
	return t.UnixNano()/1000000 - epoch
}

// TimeToMsInt64 returns the number of ms since the unix epoch as an int64
func TimeToMsInt64(t time.Time) int64 {
	return int64(t.UTC().UnixNano() / 1000000)
}
