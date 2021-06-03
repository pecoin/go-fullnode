package ethtest

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pecoin/go-fullnode/p2p"
	"github.com/pecoin/go-fullnode/pec/protocols/eth"
	"github.com/stretchr/testify/assert"
)

// TestEthProtocolNegotiation tests whether the test suite
// can negotiate the highest pec protocol in a status message exchange
func TestEthProtocolNegotiation(t *testing.T) {
	var tests = []struct {
		conn     *Conn
		caps     []p2p.Cap
		expected uint32
	}{
		{
			conn: &Conn{},
			caps: []p2p.Cap{
				{Name: "pec", Version: 63},
				{Name: "pec", Version: 64},
				{Name: "pec", Version: 65},
			},
			expected: uint32(65),
		},
		{
			conn: &Conn{},
			caps: []p2p.Cap{
				{Name: "pec", Version: 0},
				{Name: "pec", Version: 89},
				{Name: "pec", Version: 65},
			},
			expected: uint32(65),
		},
		{
			conn: &Conn{},
			caps: []p2p.Cap{
				{Name: "pec", Version: 63},
				{Name: "pec", Version: 64},
				{Name: "wrongProto", Version: 65},
			},
			expected: uint32(64),
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tt.conn.negotiateEthProtocol(tt.caps)
			assert.Equal(t, tt.expected, uint32(tt.conn.ethProtocolVersion))
		})
	}
}

// TestChain_GetHeaders tests whether the test suite can correctly
// respond to a GetBlockHeaders request from a node.
func TestChain_GetHeaders(t *testing.T) {
	chainFile, err := filepath.Abs("./testdata/chain.rlp")
	if err != nil {
		t.Fatal(err)
	}
	genesisFile, err := filepath.Abs("./testdata/genesis.json")
	if err != nil {
		t.Fatal(err)
	}

	chain, err := loadChain(chainFile, genesisFile)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req      GetBlockHeaders
		expected BlockHeaders
	}{
		{
			req: GetBlockHeaders{
				Origin: eth.HashOrNumber{
					Number: uint64(2),
				},
				Amount:  uint64(5),
				Skip:    1,
				Reverse: false,
			},
			expected: BlockHeaders{
				chain.blocks[2].Header(),
				chain.blocks[4].Header(),
				chain.blocks[6].Header(),
				chain.blocks[8].Header(),
				chain.blocks[10].Header(),
			},
		},
		{
			req: GetBlockHeaders{
				Origin: eth.HashOrNumber{
					Number: uint64(chain.Len() - 1),
				},
				Amount:  uint64(3),
				Skip:    0,
				Reverse: true,
			},
			expected: BlockHeaders{
				chain.blocks[chain.Len()-1].Header(),
				chain.blocks[chain.Len()-2].Header(),
				chain.blocks[chain.Len()-3].Header(),
			},
		},
		{
			req: GetBlockHeaders{
				Origin: eth.HashOrNumber{
					Hash: chain.Head().Hash(),
				},
				Amount:  uint64(1),
				Skip:    0,
				Reverse: false,
			},
			expected: BlockHeaders{
				chain.Head().Header(),
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			headers, err := chain.GetHeaders(tt.req)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, headers, tt.expected)
		})
	}
}
