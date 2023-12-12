package proof

import (
	"bytes"
	"math"

	"github.com/pkg/errors"
)

// MerkleTree is an implementation of a Merkle tree. Instead of copying the
// leaves to pad trees to 2^n, it just uses zero hashes.
type MerkleTree struct {
	size  uint64
	depth uint64
	nodes [][]byte
}

func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New("no leaves provided")
	}

	depth := uint64(math.Ceil(math.Log2(float64(len(data)))))
	size := uint64(math.Exp2(float64(depth)))
	nodes := make([][]byte, 2*size-1)

	// fill in the leaves
	for i, leaf := range data {
		nodes[i] = Hash(leaf)
	}

	// fill in the rest of the tree
	pos := size
	for j := depth; j > 0; j-- {
		// number of nodes at this level
		nNodes := uint64(1) << j
		for i := uint64(0); i < nNodes; i += 2 {
			nodes[pos+i/2] = Hash(append(nodes[pos-nNodes+i], nodes[pos-nNodes+i+1]...))
		}
		// advance the number of nodes we've added
		pos += nNodes / 2
	}

	return &MerkleTree{nodes: nodes, size: size, depth: depth}, nil
}

func (t *MerkleTree) Root() []byte {
	return t.nodes[len(t.nodes)-1]
}

func (t *MerkleTree) Proof(leaf []byte) ([][]byte, uint64, error) {
	index, err := t.indexOf(Hash(leaf))
	if err != nil {
		return nil, 0, err
	}

	hashes := make([][]byte, t.depth)
	pos := index
	x := index
	for i := t.depth; i > 0; i -= 1 {
		// nodes at this level
		nNodes := uint64(1) << i
		if pos%2 == 0 {
			hashes[t.depth-i] = t.nodes[pos+1]
		} else {
			hashes[t.depth-i] = t.nodes[pos-1]
		}
		// this would all be much easier if I decided to count down  the nodes instead of up,
		// but hey, we got this far. The following advances by the number of nodes at this level
		// to get to the start of the next level, then adds the new x position at that level
		pos += (nNodes - x) + x/2
		x = x / 2
	}
	return hashes, index, nil
}

func VerifyProof(leaf []byte, hashes [][]byte, index uint64, root []byte) (bool, error) {
	hash := Hash(leaf)
	for _, h := range hashes {
		if index%2 == 0 {
			hash = Hash(append(hash, h...))
		} else {
			hash = Hash(append(h, hash...))
		}
		index /= 2
	}
	return bytes.Equal(hash, root), nil
}

func (t *MerkleTree) indexOf(leaf []byte) (uint64, error) {
	for i := uint64(0); i < t.size; i++ {
		if bytes.Equal(leaf, t.nodes[i]) {
			return i, nil
		}
	}
	return 0, errors.New("not found")
}
