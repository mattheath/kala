package kala

import (
	"math/big"
	"net"
)

func MacAddressToWorkerId(mac string) (uint64, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return 0, err
	}

	workerId := new(big.Int).SetBytes([]byte(hw)).Uint64()

	return workerId, nil
}
