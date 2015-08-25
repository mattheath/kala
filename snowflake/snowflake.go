package snowflake

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mattheath/kala/util"
)

const (
	// default number of bits to use for the worker id
	defaultWorkerIdBits uint32 = 10

	// default number of bits to use for the sequence (per ms)
	defaultSequenceBits uint32 = 12

	// our bespoke epoch, as we have fewer bits for time
	defaultEpoch string = "2012-01-01T00:00:00Z"
)

var (
	ErrInvalidWorkerId  error = errors.New("Invalid worker ID - worker ID out of range")
	ErrOverflow         error = errors.New("Timestamp overflow (past end of lifespan) - unable to generate any more IDs")
	ErrSequenceOverflow error = errors.New("Sequence overflow (too many IDs generated) - unable to generate IDs for 1 millisecond")
)

// New creates a new instance of a snowflake compatible ID minter
// the worker ID must be unique otherwise ID collisions are likely to occur
func New(workerId uint32) (*Snowflake, error) {

	// initialise with the defaults, including epoch
	// 2012-01-01 00:00:00 +0000 UTC => 1325376000000
	epoch, err := time.Parse(time.RFC3339, defaultEpoch)
	if err != nil {
		return nil, err
	}

	return &Snowflake{
		workerId:     workerId,
		sequenceBits: defaultSequenceBits,
		workerIdBits: defaultWorkerIdBits,
		epoch:        util.TimeToMsInt64(epoch),
	}, nil
}

type Snowflake struct {
	sync.Mutex
	// lastTimestamp is the most recent millisecond time window encountered
	lastTimestamp int64
	// workerId - 10 bits (0 -> 1023)
	workerId uint32
	// sequence number - 12 bits, we auto-increment for same-millisecond collisions
	sequence uint32

	// Options set prior to first use
	// Time bits cannot be set, and are the remainder from our 64bit limit
	sequenceBits uint32
	workerIdBits uint32
	epoch        int64

	// Limits based on configured options
	maxSequence          uint32
	maxWorkerId          uint32
	maxAdjustedTimestamp int64

	// Once we have started minting IDs the options cannot be changed
	once        sync.Once
	initialised bool
}

// Mint a new 64bit ID based on the current time, worker id and sequence
func (sf *Snowflake) MintID() (uint64, error) {
	sf.Lock()
	defer sf.Unlock()

	// Setup locks in our configured options
	sf.once.Do(sf.setup)

	// Ensure we only mint IDs if correctly configured
	if sf.workerId > sf.maxWorkerId {
		return 0, ErrInvalidWorkerId
	}

	// Get the current timestamp in ms, adjusted to our custom epoch
	t := util.CustomTimestamp(sf.epoch, time.Now())

	// Update snowflake with this, which will increment sequence number if needed
	err := sf.update(t)
	if err != nil {
		return 0, err
	}

	// Mint a new ID
	id := sf.mintId()

	return id, nil
}

func (sf *Snowflake) Mint() (string, error) {
	id, err := sf.MintID()
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(id, 10), nil
}

// setup is called the first time we mint an ID and locks in our configured options
func (sf *Snowflake) setup() {

	// Set up limits based on configured options
	sf.maxWorkerId = (1 << sf.workerIdBits) - 1 // worker id mask
	sf.maxSequence = (1 << sf.sequenceBits) - 1 // sequence mask

	// maxAdjustedTimestamp which we can generate IDs until
	// eg. with the default worker and sequence bits we are limited to 41 bits of time
	// maxAdjustedTimestamp + epoch => 2199023255551, 2081-09-06 15:47:35 +0000 UTC (69 year range)
	sf.maxAdjustedTimestamp = -1 ^ (-1 << (64 - sf.workerIdBits - sf.sequenceBits))

	// Confirm we are initialised, so new options will be ignored
	sf.initialised = true
}

// update Snowflake with a new timestamp, causing sequence numbers to increment if necessary
func (sf *Snowflake) update(t int64) error {
	if t != sf.lastTimestamp {
		switch {
		case t < sf.lastTimestamp:
			return fmt.Errorf("Time moved backwards - unable to generate IDs for %v milliseconds", sf.lastTimestamp-t)
		case t < 0:
			return fmt.Errorf("Time is currently set before our epoch - unable to generate IDs for %v milliseconds", -1*t)
		case t > sf.maxAdjustedTimestamp:
			return ErrOverflow
		}
		sf.sequence = 0
		sf.lastTimestamp = t
	} else {
		sf.sequence = sf.sequence + 1
		if sf.sequence > sf.maxSequence {
			return ErrSequenceOverflow
		}
	}

	return nil
}

// mintId mints new 64bit IDs from the timestamp, worker ID and sequence
func (sf *Snowflake) mintId() uint64 {
	return (uint64(sf.lastTimestamp) << (sf.workerIdBits + sf.sequenceBits)) |
		(uint64(sf.workerId) << sf.sequenceBits) |
		(uint64(sf.sequence))
}
