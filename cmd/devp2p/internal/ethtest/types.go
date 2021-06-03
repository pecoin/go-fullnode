package ethtest

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"time"

	"github.com/pecoin/go-fullnode/core/types"
	"github.com/pecoin/go-fullnode/crypto"
	"github.com/pecoin/go-fullnode/internal/utesting"
	"github.com/pecoin/go-fullnode/p2p"
	"github.com/pecoin/go-fullnode/p2p/rlpx"
	"github.com/pecoin/go-fullnode/pec/protocols/eth"
	"github.com/pecoin/go-fullnode/rlp"
)

type Message interface {
	Code() int
}

type Error struct {
	err error
}

func (e *Error) Unwrap() error  { return e.err }
func (e *Error) Error() string  { return e.err.Error() }
func (e *Error) Code() int      { return -1 }
func (e *Error) String() string { return e.Error() }

func errorf(format string, args ...interface{}) *Error {
	return &Error{fmt.Errorf(format, args...)}
}

// Hello is the RLP structure of the protocol handshake.
type Hello struct {
	Version    uint64
	Name       string
	Caps       []p2p.Cap
	ListenPort uint64
	ID         []byte // secp256k1 public key

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

func (h Hello) Code() int { return 0x00 }

// Disconnect is the RLP structure for a disconnect message.
type Disconnect struct {
	Reason p2p.DiscReason
}

func (d Disconnect) Code() int { return 0x01 }

type Ping struct{}

func (p Ping) Code() int { return 0x02 }

type Pong struct{}

func (p Pong) Code() int { return 0x03 }

// Status is the network packet for the status message for pec/64 and later.
type Status eth.StatusPacket

func (s Status) Code() int { return 16 }

// NewBlockHashes is the network packet for the block announcements.
type NewBlockHashes eth.NewBlockHashesPacket

func (nbh NewBlockHashes) Code() int { return 17 }

type Transactions eth.TransactionsPacket

func (t Transactions) Code() int { return 18 }

// GetBlockHeaders represents a block header query.
type GetBlockHeaders eth.GetBlockHeadersPacket

func (g GetBlockHeaders) Code() int { return 19 }

type BlockHeaders eth.BlockHeadersPacket

func (bh BlockHeaders) Code() int { return 20 }

// GetBlockBodies represents a GetBlockBodies request
type GetBlockBodies eth.GetBlockBodiesPacket

func (gbb GetBlockBodies) Code() int { return 21 }

// BlockBodies is the network packet for block content distribution.
type BlockBodies eth.BlockBodiesPacket

func (bb BlockBodies) Code() int { return 22 }

// NewBlock is the network packet for the block propagation message.
type NewBlock eth.NewBlockPacket

func (nb NewBlock) Code() int { return 23 }

// NewPooledTransactionHashes is the network packet for the tx hash propagation message.
type NewPooledTransactionHashes eth.NewPooledTransactionHashesPacket

func (nb NewPooledTransactionHashes) Code() int { return 24 }

// Conn represents an individual connection with a peer
type Conn struct {
	*rlpx.Conn
	ourKey             *ecdsa.PrivateKey
	ethProtocolVersion uint
	caps               []p2p.Cap
}

func (c *Conn) Read() Message {
	code, rawData, _, err := c.Conn.Read()
	if err != nil {
		return errorf("could not read from connection: %v", err)
	}

	var msg Message
	switch int(code) {
	case (Hello{}).Code():
		msg = new(Hello)
	case (Ping{}).Code():
		msg = new(Ping)
	case (Pong{}).Code():
		msg = new(Pong)
	case (Disconnect{}).Code():
		msg = new(Disconnect)
	case (Status{}).Code():
		msg = new(Status)
	case (GetBlockHeaders{}).Code():
		msg = new(GetBlockHeaders)
	case (BlockHeaders{}).Code():
		msg = new(BlockHeaders)
	case (GetBlockBodies{}).Code():
		msg = new(GetBlockBodies)
	case (BlockBodies{}).Code():
		msg = new(BlockBodies)
	case (NewBlock{}).Code():
		msg = new(NewBlock)
	case (NewBlockHashes{}).Code():
		msg = new(NewBlockHashes)
	case (Transactions{}).Code():
		msg = new(Transactions)
	case (NewPooledTransactionHashes{}).Code():
		msg = new(NewPooledTransactionHashes)
	default:
		return errorf("invalid message code: %d", code)
	}
	// if message is devp2p, decode here
	if err := rlp.DecodeBytes(rawData, msg); err != nil {
		return errorf("could not rlp decode message: %v", err)
	}
	return msg
}

// ReadAndServe serves GetBlockHeaders requests while waiting
// on another message from the node.
func (c *Conn) ReadAndServe(chain *Chain, timeout time.Duration) Message {
	start := time.Now()
	for time.Since(start) < timeout {
		timeout := time.Now().Add(10 * time.Second)
		c.SetReadDeadline(timeout)
		switch msg := c.Read().(type) {
		case *Ping:
			c.Write(&Pong{})
		case *GetBlockHeaders:
			req := *msg
			headers, err := chain.GetHeaders(req)
			if err != nil {
				return errorf("could not get headers for inbound header request: %v", err)
			}

			if err := c.Write(headers); err != nil {
				return errorf("could not write to connection: %v", err)
			}
		default:
			return msg
		}
	}
	return errorf("no message received within %v", timeout)
}

func (c *Conn) Write(msg Message) error {
	// check if message is pec protocol message
	var (
		payload []byte
		err     error
	)
	payload, err = rlp.EncodeToBytes(msg)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(uint64(msg.Code()), payload)
	return err
}

// handshake checks to make sure a `HELLO` is received.
func (c *Conn) handshake(t *utesting.T) Message {
	defer c.SetDeadline(time.Time{})
	c.SetDeadline(time.Now().Add(10 * time.Second))

	// write hello to client
	pub0 := crypto.FromECDSAPub(&c.ourKey.PublicKey)[1:]
	ourHandshake := &Hello{
		Version: 5,
		Caps:    c.caps,
		ID:      pub0,
	}
	if err := c.Write(ourHandshake); err != nil {
		t.Fatalf("could not write to connection: %v", err)
	}
	// read hello from client
	switch msg := c.Read().(type) {
	case *Hello:
		// set snappy if version is at least 5
		if msg.Version >= 5 {
			c.SetSnappy(true)
		}
		c.negotiateEthProtocol(msg.Caps)
		if c.ethProtocolVersion == 0 {
			t.Fatalf("unexpected pec protocol version")
		}
		return msg
	default:
		t.Fatalf("bad handshake: %#v", msg)
		return nil
	}
}

// negotiateEthProtocol sets the Conn's pec protocol version
// to highest advertised capability from peer
func (c *Conn) negotiateEthProtocol(caps []p2p.Cap) {
	var highestEthVersion uint
	for _, capability := range caps {
		if capability.Name != "pec" {
			continue
		}
		if capability.Version > highestEthVersion && capability.Version <= 65 {
			highestEthVersion = capability.Version
		}
	}
	c.ethProtocolVersion = highestEthVersion
}

// statusExchange performs a `Status` message exchange with the given
// node.
func (c *Conn) statusExchange(t *utesting.T, chain *Chain, status *Status) Message {
	defer c.SetDeadline(time.Time{})
	c.SetDeadline(time.Now().Add(20 * time.Second))

	// read status message from client
	var message Message
loop:
	for {
		switch msg := c.Read().(type) {
		case *Status:
			if have, want := msg.Head, chain.blocks[chain.Len()-1].Hash(); have != want {
				t.Fatalf("wrong head block in status, want:  %#x (block %d) have %#x",
					want, chain.blocks[chain.Len()-1].NumberU64(), have)
			}
			if have, want := msg.TD.Cmp(chain.TD(chain.Len())), 0; have != want {
				t.Fatalf("wrong TD in status: have %v want %v", have, want)
			}
			if have, want := msg.ForkID, chain.ForkID(); !reflect.DeepEqual(have, want) {
				t.Fatalf("wrong fork ID in status: have %v, want %v", have, want)
			}
			message = msg
			break loop
		case *Disconnect:
			t.Fatalf("disconnect received: %v", msg.Reason)
		case *Ping:
			c.Write(&Pong{}) // TODO (renaynay): in the future, this should be an error
			// (PINGs should not be a response upon fresh connection)
		default:
			t.Fatalf("bad status message: %s", pretty.Sdump(msg))
		}
	}
	// make sure pec protocol version is set for negotiation
	if c.ethProtocolVersion == 0 {
		t.Fatalf("pec protocol version must be set in Conn")
	}
	if status == nil {
		// write status message to client
		status = &Status{
			ProtocolVersion: uint32(c.ethProtocolVersion),
			NetworkID:       chain.chainConfig.ChainID.Uint64(),
			TD:              chain.TD(chain.Len()),
			Head:            chain.blocks[chain.Len()-1].Hash(),
			Genesis:         chain.blocks[0].Hash(),
			ForkID:          chain.ForkID(),
		}
	}

	if err := c.Write(status); err != nil {
		t.Fatalf("could not write to connection: %v", err)
	}

	return message
}

// waitForBlock waits for confirmation from the client that it has
// imported the given block.
func (c *Conn) waitForBlock(block *types.Block) error {
	defer c.SetReadDeadline(time.Time{})

	timeout := time.Now().Add(20 * time.Second)
	c.SetReadDeadline(timeout)
	for {
		req := &GetBlockHeaders{Origin: eth.HashOrNumber{Hash: block.Hash()}, Amount: 1}
		if err := c.Write(req); err != nil {
			return err
		}
		switch msg := c.Read().(type) {
		case *BlockHeaders:
			if len(*msg) > 0 {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		default:
			return fmt.Errorf("invalid message: %s", pretty.Sdump(msg))
		}
	}
}
