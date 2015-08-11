package bigflake

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/mattheath/kala/util"
)

const (
	// default number of bits to use for the worker id
	defaultWorkerIdBits uint32 = 48

	// default number of bits to use for the sequence (per ms)
	defaultSequenceBits uint32 = 16
)

var (
	ErrInvalidWorkerId  error = errors.New("Invalid worker ID - worker ID out of range")
	ErrOverflow         error = errors.New("Timestamp overflow (past end of lifespan) - unable to generate any more IDs")
	ErrSequenceOverflow error = errors.New("Sequence overflow (too many IDs generated) - unable to generate IDs for 1 millisecond")
)

// New initialises a Bigflake minter, with a default configuration
// This can be configured using Options
func New(workerId uint64) (*Bigflake, error) {
	return &Bigflake{
		workerId:     int64(workerId),
		sequenceBits: defaultSequenceBits,
		workerIdBits: defaultWorkerIdBits,
		epoch:        0, // default unix epoch
	}, nil
}

type Bigflake struct {
	sync.Mutex

	// lastTimestamp is the most recent millisecond time window encountered
	lastTimestamp int64

	// workerId - 48 bits
	workerId int64
	// sequence number - 16 bits
	// we auto-increment for same-millisecond collisions
	sequence int64

	// Options set prior to first use
	// Time bits cannot be set, and are the remainder from our bit limit
	sequenceBits uint32
	workerIdBits uint32
	epoch        int64

	// Limits based on configured options
	maxSequence          int64
	maxWorkerId          int64
	maxAdjustedTimestamp int64

	// Once we have started minting IDs the options cannot be changed
	once        sync.Once
	initialised bool
}

// Mint a new 128bit ID based on the current time, worker id and sequence
func (bf *Bigflake) Mint() (*BigflakeId, error) {
	bf.Lock()
	defer bf.Unlock()

	// Setup locks in our configured options
	bf.once.Do(bf.setup)

	// Ensure we only mint IDs if correctly configured
	if bf.workerId > bf.maxWorkerId {
		return nil, ErrInvalidWorkerId
	}

	// Get the current timestamp in ms
	// @todo generalise to allow custom epoch
	t := util.TimeToMsInt64(time.Now())

	// Update bigflake with this, which Mawill increment sequence number if needed
	err := bf.update(t)
	if err != nil {
		return nil, err
	}

	// Mint a new ID
	id := mintId(bf.lastTimestamp, bf.workerId, bf.sequence, bf.workerIdBits, bf.sequenceBits)
	bfId := &BigflakeId{
		id: id,
	}

	return bfId, nil
}

// setup is called the first time we mint an ID and locks in our configured options
func (bf *Bigflake) setup() {

	// Set up limits based on configured options
	bf.maxWorkerId = (1 << bf.workerIdBits) - 1 // worker id mask
	bf.maxSequence = (1 << bf.sequenceBits) - 1 // sequence mask

	// maxAdjustedTimestamp which we can generate IDs until
	// eg. with the default worker and sequence bits we are limited to 41 bits of time
	// maxAdjustedTimestamp + epoch => 2199023255551, 2081-09-06 15:47:35 +0000 UTC (69 year range)
	// bf.maxAdjustedTimestamp = -1 ^ (-1 << (64 - sf.workerIdBits - sf.sequenceBits))

	// Confirm we are initialised, so new options will be ignored
	bf.initialised = true
}

// update Bigflake with a new timestamp, causing sequence numbers to increment if necessary
func (bf *Bigflake) update(t int64) error {
	if t != bf.lastTimestamp {
		// fmt.Println("Time not equal")

		switch {
		case t < bf.lastTimestamp:
			return fmt.Errorf("Time moved backwards - unable to generate IDs for %v milliseconds", bf.lastTimestamp-t)
		case t < 0:
			return fmt.Errorf("Time is currently set before our epoch - unable to generate IDs for %v milliseconds", -1*t)
			// case t > bf.maxAdjustedTimestamp:
			// 	return ErrOverflow
		}

		// Reset sequence as we're in a new ms
		bf.sequence = 0
		bf.lastTimestamp = t
	}

	// fmt.Printf("Bigflake: %#v\n\n", bf)

	// Increment sequence for this ms
	bf.sequence = bf.sequence + 1
	if bf.sequence > bf.maxSequence {
		return ErrSequenceOverflow
	}

	return nil
}

// MintId mints new 128bit IDs from the timestamp, worker ID and sequence,
// this should only be used directly for testing
func MintId(timestamp, workerid, sequence int64) *big.Int {
	return mintId(timestamp, workerid, sequence, defaultWorkerIdBits, defaultSequenceBits)
}

func mintId(timestamp, workerid, sequence int64,
	workerIdBits, sequenceIdBits uint32) *big.Int {

	// Time is the most significant bits
	// Shift by the number of worker and sequence bits
	id := big.NewInt(timestamp)
	id = id.Lsh(id, uint(workerIdBits+sequenceIdBits))

	// Shift the worker ID by the number of sequence bits
	bigW := big.NewInt(workerid)
	bigW = bigW.Lsh(bigW, uint(sequenceIdBits))

	// Sequence doesn't need shifting
	bigS := big.NewInt(sequence)

	// Combine components with bitwise OR
	id = id.Or(id, bigW)
	id = id.Or(id, bigS)

	return id
}

func ParseId(id *big.Int) (timestamp, workerid, sequence int64) {
	bigS := big.NewInt(0)
	bigW := big.NewInt(0)

	bigS.And(id, big.NewInt((1<<defaultSequenceBits)-1))
	id.Rsh(id, uint(defaultSequenceBits))
	bigW.And(id, big.NewInt((1<<defaultWorkerIdBits)-1))
	id.Rsh(id, uint(defaultWorkerIdBits))

	return id.Int64(), bigW.Int64(), bigS.Int64()
}
