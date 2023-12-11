package api

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/model"
	"github.com/scottrmalley/p2p-file-challenge/proof"
)

var ErrFileSetCorrupted = errors.New("file set corrupted")

type broadcaster interface {
	Broadcast(ctx context.Context, files []model.File) error
}

type processor interface {
	ProcessFiles(setId string, files []model.File) ([]byte, error)
}

type persistence interface {
	File(setId string, index int) (model.File, error)
	FileIds(setId string) ([]string, error)
}

type Service struct {
	logger      zerolog.Logger
	broadcaster broadcaster
	repo        persistence
	processor   processor
}

func NewService(logger zerolog.Logger, broadcaster broadcaster, repo persistence, process processor) *Service {
	return &Service{
		logger:      logger,
		broadcaster: broadcaster,
		repo:        repo,
		processor:   process,
	}
}

func (s *Service) Files(setId uuid.UUID, files [][]byte) (string, error) {
	var result []model.File
	fileStream := make(chan model.File, len(files))
	for i, file := range files {
		result = append(
			result, model.File{
				Metadata: model.FileMetadata{
					SetId:      setId.String(),
					SetCount:   len(files),
					FileNumber: i,
				},
				Contents: file,
			},
		)
		fileStream <- result[i]
	}

	err := s.broadcaster.Broadcast(context.Background(), result)
	if err != nil {
		return "", err
	}

	root, err := s.processor.ProcessFiles(setId.String(), result)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(root), nil
}

func (s *Service) File(setId uuid.UUID, index int) ([]byte, [][]byte, uint64, error) {
	file, err := s.repo.File(setId.String(), index)
	if err != nil {
		return nil, nil, 0, err
	}
	ids, err := s.repo.FileIds(setId.String())
	if err != nil {
		return nil, nil, 0, err
	}

	if len(ids) != file.Metadata.SetCount {
		return nil, nil, 0, ErrFileSetCorrupted
	}

	if file.Metadata.Id != hexutil.Encode(crypto.Keccak256(file.Contents)) {
		return nil, nil, 0, ErrFileSetCorrupted
	}

	leaves := make([][]byte, len(ids))
	for i, id := range ids {
		leaf, err := hexutil.Decode(id)
		if err != nil {
			return nil, nil, 0, err
		}
		leaves[i] = leaf
	}

	path, position, err := proof.Proof(
		leaves,
		crypto.Keccak256(file.Contents),
	)

	return file.Contents, path, position, err
}
