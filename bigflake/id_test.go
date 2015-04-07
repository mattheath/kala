// Tests encoding/decoding of IDs

package bigflake

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var idTestCases = []struct {
	base10 string
	uuid   string
	base62 string
}{
	// now, with mac address worker id
	{"26344968761766525548891622211585", "0000014c-852f-65e6-8036-bcdb64160001", "8ucl7ptu4YVHsRigKn"},
	{"26344968761766525548891622211586", "0000014c-852f-65e6-8036-bcdb64160002", "8ucl7ptu4YVHsRigKo"},
	{"26344968761766525548891622211587", "0000014c-852f-65e6-8036-bcdb64160003", "8ucl7ptu4YVHsRigKp"},
	{"26344968761766525548891622211588", "0000014c-852f-65e6-8036-bcdb64160004", "8ucl7ptu4YVHsRigKq"},
	{"26344968761766525548891622211589", "0000014c-852f-65e6-8036-bcdb64160005", "8ucl7ptu4YVHsRigKr"},

	// 10 years in future
	{"32167119573924573840679378485250", "00000196-0191-ac58-8036-bcdb64160002", "AskiQUWuearuAEjYqg"},
	{"32167119573924573840679378485251", "00000196-0191-ac58-8036-bcdb64160003", "AskiQUWuearuAEjYqh"},
	{"32167119573924573840679378485252", "00000196-0191-ac58-8036-bcdb64160004", "AskiQUWuearuAEjYqi"},

	// low worker id
	{"26344973791391185674980777656326", "0000014c-8533-8ef7-0000-0000000a0006", "8uclWyrUsELNlbIc9W"},
	{"26344973791391185674980777656327", "0000014c-8533-8ef7-0000-0000000a0007", "8uclWyrUsELNlbIc9X"},
	{"26344973791391185674980777656328", "0000014c-8533-8ef7-0000-0000000a0008", "8uclWyrUsELNlbIc9Y"},

	// near max 64bit time
	{"170141183460469231602560095199917899778", "7fffffff-ffff-fff9-0000-0000000a0002", "3tX16dB2jpqOE7aKJrVcQc"},
	{"170141183460469231602560095199917899779", "7fffffff-ffff-fff9-0000-0000000a0003", "3tX16dB2jpqOE7aKJrVcQd"},
}

// Test marshaling back and forth between string and int
// Pretty much a sanity check for further tests
func TestStringMarshal(t *testing.T) {
	for _, tc := range idTestCases {
		i := new(big.Int)
		i.SetString(tc.base10, 10)
		assert.Equal(t, tc.base10, i.String())
	}
}

func TestUuidMarshal(t *testing.T) {
	for _, tc := range idTestCases {
		i := new(big.Int)
		i.SetString(tc.base10, 10)
		id := &BigflakeId{
			id: i,
		}
		assert.Equal(t, tc.uuid, id.Uuid())
	}
}

func TestBase62Marshal(t *testing.T) {
	for _, tc := range idTestCases {
		i := new(big.Int)
		i.SetString(tc.base10, 10)
		id := &BigflakeId{
			id: i,
		}
		assert.Equal(t, tc.base62, id.Base62())
	}
}

func TestParseUuid(t *testing.T) {
	for _, tc := range idTestCases {
		// marshal to uuid first
		i := new(big.Int)
		i.SetString(tc.base10, 10)
		id := &BigflakeId{
			id: i,
		}
		s := id.Uuid()

		// then parse and compare back to original
		newBf, err := ParseUuid(s)
		require.NoError(t, err)
		assert.Equal(t, tc.base10, newBf.String())
		t.Logf("base10: %s | uuid: %s | base10: %s", tc.base10, s, newBf.String())
	}
}
