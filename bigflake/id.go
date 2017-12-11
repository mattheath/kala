package bigflake

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/mattheath/base62"
)

// NewId creates a BigflakeId from a big.Int
func NewId(id *big.Int) *BigflakeId {
	return &BigflakeId{id}
}

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

// Base62WithPadding returns a base62 encoded id with left padding
func (bf *BigflakeId) Base62WithPadding(minlen int) string {
	e := base62.NewStdEncoding().Option(base62.Padding(minlen))

	return e.EncodeBigInt(bf.id)
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

// UUID Parsing code based on github.com/nu7hatch/gouuid
// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch>

// Pattern used to parse hex string representation of the UUID.
var uuidRegexp = regexp.MustCompile("^(urn\\:uuid\\:)?\\{?([a-z0-9]{8})-([a-z0-9]{4})-" +
	"([a-z0-9]{4})-([a-z0-9]{4})-([a-z0-9]{12})\\}?$")

// ParseUuid into a BigflakeId
func ParseUuid(s string) (bf *BigflakeId, err error) {
	md := uuidRegexp.FindStringSubmatch(s)
	if md == nil {
		err = errors.New("Invalid UUID string")
		return
	}
	hash := md[2] + md[3] + md[4] + md[5] + md[6]
	b, err := hex.DecodeString(hash)
	if err != nil {
		return
	}

	id := new(big.Int)
	id.SetBytes(b)
	bf = &BigflakeId{
		id: id,
	}
	return
}
