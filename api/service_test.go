package api

import (
	"fmt"
	"io"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/scottrmalley/p2p-file-sharing/proof"
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
		"it should save a set", func(t *testing.T) {
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
			for i, file := range testFiles {
				_, err := service.SaveFile(
					setId,
					i,
					len(testFiles),
					file,
				)
				s.NoError(err)
			}

			files, err := s.repo.Files(setId.String())
			s.NoError(err)
			s.Len(files, len(testFiles))
			for i, file := range files {
				s.Equal(testFiles[i], file)
			}
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
			for i, file := range testFiles {
				_, err := service.SaveFile(
					setId,
					i,
					len(testFiles),
					file,
				)
				s.NoError(err)
			}

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
			fmt.Printf("testFiles: %s\n", hexutil.Encode(testFiles[0]))
			setId := uuid.New()
			for i, file := range testFiles {
				_, err := service.SaveFile(
					setId,
					i,
					len(testFiles),
					file,
				)
				s.NoError(err)
			}

			expectedRoot, err := proof.Root(testFiles)
			s.NoError(err)

			file, hashes, index, err := service.File(setId, 0)
			s.NoError(err)

			verified, err := proof.Verify(file, hashes, index, expectedRoot)
			s.NoError(err)
			s.True(verified)
		},
	)

}
