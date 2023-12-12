package proof

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// These are just wrappers so we have a centralized place to change the
// has or encoding functions if we need to

func Hash(data []byte) []byte {
	return crypto.Keccak256(data)
}

func Encode(data []byte) string {
	return hexutil.Encode(data)
}

func Decode(data string) ([]byte, error) {
	return hexutil.Decode(data)
}
