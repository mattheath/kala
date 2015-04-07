package bigflake

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/mattheath/base62"
)

// BigflakeId represents a globally unique ID
type BigflakeId struct {
	id *big.Int
}

// String returns the raw id as a string
func (bf *BigflakeId) String() string {
	return bf.id.String()
}

// BinaryString returns a padded 128bit binary number formatted as a string
func (bf *BigflakeId) BinaryString() string {
	return fmt.Sprintf("%0128b", bf.id)
}

// Base62 returns a base62 encoded version
func (bf *BigflakeId) Base62() string {
	return base62.EncodeBigInt(bf.id)
}

// Uuid returns the id encoded in UUID format
func (bf *BigflakeId) Uuid() string {
	b := bf.id.Bytes()

	// Pad numbers less than the full 128bit (16 byte) width
	if len(b) < 16 {
		pad := 16 - len(b)
		padslice := make([]byte, pad)
		// prepend to bytes
		b = append(padslice, b...)
	}

	// Return hex formatted with delimiters
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:])
}

// Raw returns a raw 128bit integer
func (bf *BigflakeId) Raw() *big.Int {
	return bf.id
}
