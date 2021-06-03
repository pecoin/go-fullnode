// Contains all the wrappers from the accounts package to support client side enode
// management on mobile platforms.

package geth

import (
	"errors"

	"github.com/pecoin/go-fullnode/p2p/enode"
)

// Enode represents a host on the network.
type Enode struct {
	node *enode.Node
}

// NewEnode parses a node designator.
//
// There are two basic forms of node designators
//   - incomplete nodes, which only have the public key (node ID)
//   - complete nodes, which contain the public key and IP/Port information
//
// For incomplete nodes, the designator must look like one of these
//
//    enode://<hex node id>
//    <hex node id>
//
// For complete nodes, the node ID is encoded in the username portion
// of the URL, separated from the host by an @ sign. The hostname can
// only be given as an IP address, DNS domain names are not allowed.
// The port in the host name section is the TCP listening port. If the
// TCP and UDP (discovery) ports differ, the UDP port is specified as
// query parameter "discport".
//
// In the following example, the node URL describes
// a node with IP address 10.3.58.6, TCP listening port 21861
// and UDP discovery port 21861.
//
//    enode://<hex node id>@10.3.58.6:21861?discport=0
func NewEnode(rawurl string) (*Enode, error) {
	node, err := enode.Parse(enode.ValidSchemes, rawurl)
	if err != nil {
		return nil, err
	}
	return &Enode{node}, nil
}

// Enodes represents a slice of accounts.
type Enodes struct{ nodes []*enode.Node }

// NewEnodes creates a slice of uninitialized enodes.
func NewEnodes(size int) *Enodes {
	return &Enodes{
		nodes: make([]*enode.Node, size),
	}
}

// NewEnodesEmpty creates an empty slice of Enode values.
func NewEnodesEmpty() *Enodes {
	return NewEnodes(0)
}

// Size returns the number of enodes in the slice.
func (e *Enodes) Size() int {
	return len(e.nodes)
}

// Get returns the enode at the given index from the slice.
func (e *Enodes) Get(index int) (enode *Enode, _ error) {
	if index < 0 || index >= len(e.nodes) {
		return nil, errors.New("index out of bounds")
	}
	return &Enode{e.nodes[index]}, nil
}

// Set sets the enode at the given index in the slice.
func (e *Enodes) Set(index int, enode *Enode) error {
	if index < 0 || index >= len(e.nodes) {
		return errors.New("index out of bounds")
	}
	e.nodes[index] = enode.node
	return nil
}

// Append adds a new enode element to the end of the slice.
func (e *Enodes) Append(enode *Enode) {
	e.nodes = append(e.nodes, enode.node)
}
