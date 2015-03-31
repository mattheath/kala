package kala

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	workerIdBits uint32 = 10                        // worker id
	maxWorkerId  uint32 = -1 ^ (-1 << workerIdBits) // worker id mask
	sequenceBits uint32 = 12                        // sequence
	maxSequence  uint32 = -1 ^ (-1 << sequenceBits) // sequence mask

	// maxAdjustedTimestamp which we can generate IDs to, as we are limited to 41 bits
	// maxAdjustedTimestamp + epoch => 2081-09-06 15:47:35 +0000 UTC (69 year range)
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

// NewSnowflake creates a new instance of a snowflake compatible ID minter
// the worker ID must be unique otherwise ID collisions are likely to occur
func NewSnowflake(workerId uint32) (Minter, error) {
	if workerId < 0 || workerId > maxWorkerId {
		return nil, ErrInvalidWorkerId
	}
	return &goFlake{workerId: workerId}, nil
}

type goFlake struct {
	sync.Mutex
	// lastTimestamp is the most recent millisecond time window encountered
	lastTimestamp int64
	// workerId - 10 bits (0 -> 1023)
	workerId uint32
	// sequence number - 12 bits, we auto-increment for same-millisecond collisions
	sequence uint32
}

// Mint a new 64bit ID based on the current time, worker id and sequence
func (gf *goFlake) Mint() (string, error) {
	gf.Lock()
	defer gf.Unlock()

	// Get the current timestamp in ms, adjusted to our custom epoch
	t := customTimestamp(time.Now())

	// Update goflake with this, which will increment sequence number if needed
	err := gf.update(t)
	if err != nil {
		return "", err
	}

	// Mint a new ID
	id := gf.mintId()

	return strconv.FormatUint(id, 10), nil
}

// update GoFlake with a new timestamp, causing sequence numbers to increment if necessary
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

// mintId mints new 64bit IDs from the timestamp, worker ID and sequence
func (gf *goFlake) mintId() uint64 {
	return (uint64(gf.lastTimestamp) << (workerIdBits + sequenceBits)) |
		(uint64(gf.workerId) << sequenceBits) |
		(uint64(gf.sequence))
}

// customTimestamp takes a timestamp and adjusts it to our custom epoch
func customTimestamp(t time.Time) int64 {
	return t.UnixNano()/1000000 - epoch
}
