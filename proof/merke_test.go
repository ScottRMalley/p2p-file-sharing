package proof

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProofTestSuite is the test suite for the proof package.
// For this implementation, the tests are not very significant, as we are
// merely wrapping the go-merkletree package, however it is useful to have
// tests anyway in case we swap out the package and want to ensure the same
// functionality.
type ProofTestSuite struct {
	suite.Suite
}

func TestProof(t *testing.T) {
	suite.Run(t, new(ProofTestSuite))
}

func (s *ProofTestSuite) TestProof() {
	t := s.T()
	t.Run(
		"it should return the root", func(t *testing.T) {
			data := [][]byte{
				[]byte("foo"),
				[]byte("bar"),
			}
			expectedRoot := crypto.Keccak256(
				crypto.Keccak256([]byte("foo")),
				crypto.Keccak256([]byte("bar")),
			)

			root, err := Root(data)
			require.NoError(t, err)
			require.NotNil(t, root)

			require.Equal(t, expectedRoot, root)
		},
	)

	t.Run(
		"it should return the proof", func(t *testing.T) {
			data := [][]byte{
				[]byte("foo"),
				[]byte("bar"),
			}
			leaf := []byte("foo")

			expectedProof := [][]byte{
				crypto.Keccak256([]byte("bar")),
			}

			proof, index, err := Proof(data, leaf)
			require.NoError(t, err)
			require.NotNil(t, proof)

			require.Equal(t, uint64(0), index)
			require.Equal(t, expectedProof, proof)
		},
	)

	t.Run(
		"it should verify the proof", func(t *testing.T) {
			leaf := []byte("foo")
			proof := [][]byte{
				crypto.Keccak256([]byte("bar")),
			}
			root := crypto.Keccak256(
				crypto.Keccak256([]byte("foo")),
				crypto.Keccak256([]byte("bar")),
			)

			valid, err := Verify(leaf, proof, uint64(0), root)
			require.NoError(t, err)
			require.True(t, valid)
		},
	)

	t.Run(
		"it should verify a much longer proof", func(t *testing.T) {
			data := [][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("dachschund"),
				[]byte("corgie"),
				[]byte("poodle"),
				[]byte("labrador"),
				[]byte("husky"),
				[]byte("pug"),
				[]byte("beagle"),
				[]byte("boxer"),
			}
			leaf := []byte("corgie")

			proof, index, err := Proof(data, leaf)
			require.NoError(t, err)

			root, err := Root(data)
			require.NoError(t, err)

			valid, err := Verify(leaf, proof, index, root)
			require.NoError(t, err)
			require.True(t, valid)
		},
	)
}
