package api

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/proof"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite
	repo *persistenceMock
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.repo = newPersistenceMock()
}

func (s *ServiceTestSuite) TestFiles() {
	t := s.T()
	t.Run(
		"it should return a set id", func(t *testing.T) {
			service := NewService(
				zerolog.New(io.Discard),
				s.repo,
				s.repo,
			)
			testFiles := [][]byte{
				[]byte("file1"),
				[]byte("file2"),
			}
			root, err := service.SaveFiles(
				uuid.New(),
				testFiles,
			)
			s.NoError(err)
			s.NotEmpty(root)

			expectedRoot, err := proof.Root(testFiles)
			s.NoError(err)
			s.Equal(hexutil.Encode(expectedRoot), root)
		},
	)
	t.Run(
		"it should retrieve the file", func(t *testing.T) {
			service := NewService(
				zerolog.New(io.Discard),
				s.repo,
				s.repo,
			)
			testFiles := [][]byte{
				[]byte("file1"),
				[]byte("file2"),
			}
			setId := uuid.New()
			root, err := service.SaveFiles(
				setId,
				testFiles,
			)
			s.NoError(err)
			s.NotEmpty(root)

			expectedRoot, err := proof.Root(testFiles)
			s.NoError(err)
			s.Equal(hexutil.Encode(expectedRoot), root)

			file, _, _, err := service.File(setId, 0)
			s.NoError(err)
			s.Equal(testFiles[0], file)
		},
	)

	t.Run(
		"it should return the correct proof", func(t *testing.T) {
			service := NewService(
				zerolog.New(io.Discard),
				s.repo,
				s.repo,
			)
			testFiles := [][]byte{
				[]byte("file1"),
				[]byte("file2"),
			}
			setId := uuid.New()
			root, err := service.SaveFiles(
				setId,
				testFiles,
			)
			s.NoError(err)
			s.NotEmpty(root)

			expectedRoot, err := proof.Root(testFiles)
			s.NoError(err)
			s.Equal(hexutil.Encode(expectedRoot), root)

			file, hashes, index, err := service.File(setId, 0)
			s.NoError(err)

			verified, err := proof.Verify(file, hashes, index, expectedRoot)
			s.NoError(err)
			s.True(verified)
		},
	)

}
