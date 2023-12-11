package protocol

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/scottrmalley/p2p-file-challenge/model"
	"github.com/scottrmalley/p2p-file-challenge/proof"
)

var ErrSetIncomplete = errors.New("set is not complete")

type persistence interface {
	SaveFile(string, model.File) error
}

type Processor struct {
	repo    persistence
	pending pendingFileCache
}

func NewProcessor(repo persistence) *Processor {
	return &Processor{
		repo: repo,
		pending: pendingFileCache{
			pendingFiles: make(map[string][]model.File),
		},
	}
}

func (p *Processor) ProcessFile(file model.File) error {
	fileId := hexutil.Encode(crypto.Keccak256(file.Contents))

	// first save the file
	err := p.repo.SaveFile(fileId, file)
	if err != nil {
		return err
	}

	// add file to pending cache
	p.pending.AddFile(file)

	return nil
}

func (p *Processor) CompleteSet(setId string) ([]byte, error) {
	if !p.pending.IsComplete(setId) {
		return nil, ErrSetIncomplete
	}

	files := p.pending.Remove(setId)

	return proof.Root(
		func() [][]byte {
			var result [][]byte
			for _, file := range files {
				result = append(result, file.Contents)
			}
			return result
		}(),
	)
}

func (p *Processor) ProcessFiles(setId string, files []model.File) ([]byte, error) {
	for _, file := range files {
		if file.Metadata.SetId != setId {
			return nil, errors.New("file set id does not match")
		}
		err := p.ProcessFile(file)
		if err != nil {
			return nil, err
		}
	}

	return p.CompleteSet(setId)
}
