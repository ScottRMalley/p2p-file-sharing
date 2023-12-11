package repository

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/model"
	"gorm.io/gorm"
)

type Files struct {
	logger zerolog.Logger
	db     *gorm.DB
}

type fileModel struct {
	gorm.Model
	SetId    string
	FileHash string
	Contents []byte

	SetCount,
	FileNumber int
}

type fileIdModel struct {
	FileHash string
}

func NewFiles(logger zerolog.Logger, db *gorm.DB) *Files {
	return &Files{
		logger: logger,
		db:     db,
	}
}

func (r *Files) Migrate() error {
	r.logger.Info().Msg("applying file table migrations")

	if err := r.db.AutoMigrate(&fileModel{}); err != nil {
		return errors.Wrap(err, "migration for fileModel failed")
	}
	return nil
}

func (r *Files) SaveFile(file model.File) error {
	hash := hexutil.Encode(crypto.Keccak256(file.Contents))
	result := r.db.Create(
		&fileModel{
			SetId:      file.Metadata.SetId,
			SetCount:   file.Metadata.SetCount,
			FileHash:   hash,
			FileNumber: file.Metadata.FileNumber,
			Contents:   file.Contents,
		},
	)

	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to save file")
	}

	if result.RowsAffected != 1 {
		return errors.New("failed to save file")
	}

	return nil
}

func (r *Files) File(setId string, index int) (model.File, error) {
	var file fileModel
	result := r.db.Where("set_id = ? AND file_number = ?", setId, index).First(&file)
	if result.Error != nil {
		return model.File{}, errors.Wrap(result.Error, "failed to get file")
	}
	return model.File{
		Metadata: model.FileMetadata{
			SetId:      file.SetId,
			SetCount:   file.SetCount,
			FileNumber: file.FileNumber,
		},
		Contents: file.Contents,
	}, nil
}

func (r *Files) Files(setId string) ([][]byte, error) {
	var files []fileModel

	// using to fileIdModel automatically selects only the fileId column
	result := r.db.Where("set_id = ?", setId).Find(&files).Order("file_number ASC")
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to get file hashes")
	}

	hashes := make([][]byte, len(files))
	for i, file := range files {
		hashes[i] = file.Contents
	}

	return hashes, nil
}
