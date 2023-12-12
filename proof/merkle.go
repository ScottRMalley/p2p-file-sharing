package proof

// Some wrapper functions to not have to deal with trees elsewhere
// in the application

func Root(data [][]byte) ([]byte, error) {
	if tree, err := NewMerkleTree(data); err != nil {
		return nil, err
	} else {
		return tree.Root(), nil
	}
}

func Proof(data [][]byte, leaf []byte) ([][]byte, uint64, error) {
	tree, err := NewMerkleTree(data)
	if err != nil {
		return nil, 0, err
	}
	proof, pos, err := tree.Proof(leaf)
	if err != nil {
		return nil, 0, err
	}
	return proof, pos, nil
}

func Verify(leaf []byte, proof [][]byte, index uint64, root []byte) (bool, error) {
	return VerifyProof(leaf, proof, index, root)
}
