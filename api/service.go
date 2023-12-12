package api

import (
	"context"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/scottrmalley/p2p-file-sharing/model"
	"github.com/scottrmalley/p2p-file-sharing/proof"
)

var ErrFileSetIncomplete = errors.New("file set incomplete")

type Writer interface {
	Write(ctx context.Context, file model.File) error
}
type persistence interface {
	SaveFile(file model.File) error
	File(setId string, index int) (model.File, error)
	Files(setId string) ([][]byte, error)
}

type Service struct {
	logger zerolog.Logger
	writer Writer
	repo   persistence
}

func NewService(logger zerolog.Logger, writer Writer, repo persistence) *Service {
	return &Service{
		logger: logger,
		writer: writer,
		repo:   repo,
	}
}

func (s *Service) SaveFile(setId uuid.UUID, index, setCount int, file []byte) (string, error) {
	f := model.File{
		Metadata: model.FileMetadata{
			SetId:      setId.String(),
			SetCount:   setCount,
			FileNumber: index,
		},
		Contents: file,
	}
	err := s.writer.Write(context.Background(), f)
	if err != nil {
		return "", err
	}

	if err = s.repo.SaveFile(f); err != nil {
		return "", err
	}

	return hexutil.Encode(crypto.Keccak256(file)), nil
}

func (s *Service) File(setId uuid.UUID, index int) ([]byte, [][]byte, uint64, error) {
	file, err := s.repo.File(setId.String(), index)
	if err != nil {
		return nil, nil, 0, err
	}
	files, err := s.repo.Files(setId.String())
	if err != nil {
		return nil, nil, 0, err
	}

	if len(files) != file.Metadata.SetCount {
		return nil, nil, 0, ErrFileSetIncomplete
	}

	path, position, err := proof.Proof(
		files,
		file.Contents,
	)

	return file.Contents, path, position, nil
}
