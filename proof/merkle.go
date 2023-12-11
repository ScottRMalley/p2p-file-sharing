package proof

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/wealdtech/go-merkletree"
)

type hasher struct {
}

// Hash Just using my favorite hash function here as I am used to seeing 0x... strings
// If it's not good then Ethereum has bigger problems than I do
func (h *hasher) Hash(data []byte) []byte {
	return crypto.Keccak256(data)
}

func Root(data [][]byte) ([]byte, error) {
	if tree, err := merkletree.NewUsing(data, &hasher{}, nil); err != nil {
		return nil, err
	} else {
		return tree.Root(), nil
	}
}

func Proof(data [][]byte, leaf []byte) ([][]byte, uint64, error) {
	tree, err := merkletree.NewUsing(data, &hasher{}, nil)
	if err != nil {
		return nil, 0, err
	}
	proof, err := tree.GenerateProof(leaf)
	if err != nil {
		return nil, 0, err
	}
	return proof.Hashes, proof.Index, nil
}

func Verify(leaf []byte, proof [][]byte, index uint64, root []byte) (bool, error) {
	return merkletree.VerifyProofUsing(
		leaf,
		&merkletree.Proof{Hashes: proof, Index: index},
		root,
		&hasher{},
		nil,
	)
}
