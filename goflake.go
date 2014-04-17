package goflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	workerIdBits uint32 = 10                        // worker id
	maxWorkerId  uint32 = -1 ^ (-1 << workerIdBits) // worker id mask
	sequenceBits uint32 = 12                        // sequence
	maxSequence  uint32 = -1 ^ (-1 << sequenceBits) // sequence mask

	// maxAdjustedTimestamp which we can generate IDs to, as we are limited to 41 bits
	// maxAdjustedTimestamp + epoch => 2081-09-06 15:47:35 +0000 UTC
	maxAdjustedTimestamp int64 = 2199023255551
)

var (
	ErrOverflow         error = errors.New(fmt.Sprintf("Timestamp overflow (past end of lifespan) - unable to generate any more IDs"))
	ErrInvalidWorkerId  error = errors.New(fmt.Sprintf("Invalid worker ID - worker ID must be between 0 and %v", maxWorkerId))
	ErrSequenceOverflow error = errors.New(fmt.Sprintf("Sequence overflow (too many IDs generated) - unable to generate IDs for 1 millisecond"))

	// epoch as UTC millisecond timestamp
	// 2012-01-01 00:00:00 +0000 UTC => 1325376000000
	epoch int64 = int64(time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1000000)
)

type goFlake struct {
	sync.Mutex
	// lastTimestamp is the most recent millisecond time window encountered
	lastTimestamp int64
	// workerId - 10 bits (0 -> 1023)
	workerId uint32
	// sequence number - 12 bits, we auto-increment for same-millisecond collisions
	sequence uint32
}

func (gf *goFlake) Generate() (uint64, error) {
	gf.Lock()
	defer gf.Unlock()

	t := customTimestamp(time.Now())

	err := gf.update(t)
	if err != nil {
		return 0, err
	}

	id := gf.mintId()
	return id, nil
}

func (gf *goFlake) update(t int64) error {
	if t != gf.lastTimestamp {
		switch {
		case t < gf.lastTimestamp:
			return fmt.Errorf("Time moved backwards - unable to generate IDs for %v milliseconds", gf.lastTimestamp-t)
		case t < 0:
			return fmt.Errorf("Time is currently set before our epoch - unable to generate IDs for %v milliseconds", -1*t)
		case t > maxAdjustedTimestamp:
			return ErrOverflow
		}
		gf.sequence = 0
		gf.lastTimestamp = t
	} else {
		gf.sequence = gf.sequence + 1
		if gf.sequence > maxSequence {
			return ErrSequenceOverflow
		}
	}

	return nil
}

func (gf *goFlake) mintId() uint64 {
	return (uint64(gf.lastTimestamp) << (workerIdBits + sequenceBits)) |
		(uint64(gf.workerId) << sequenceBits) |
		(uint64(gf.sequence))
}

func New(workerId uint32) (*goFlake, error) {
	if workerId < 0 || workerId > maxWorkerId {
		return nil, ErrInvalidWorkerId
	}
	return &goFlake{workerId: workerId}, nil
}

func customTimestamp(t time.Time) int64 {
	return t.UnixNano()/1000000 - epoch
}
