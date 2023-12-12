package repository

import (
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/scottrmalley/p2p-file-sharing/model"
	"github.com/scottrmalley/p2p-file-sharing/proof"
)

type Files struct {
	logger zerolog.Logger
	mu     sync.Mutex
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

type fileModelstruct []fileModel

func (f fileModelstruct) Len() int {
	return len(f)
}

func (f fileModelstruct) Less(i, j int) bool {
	return f[i].FileNumber < f[j].FileNumber
}

func (f fileModelstruct) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
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
	r.mu.Lock()
	defer r.mu.Unlock()
	hash := proof.Encode(proof.Hash(file.Contents))
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
		return nil, errors.Wrap(result.Error, "failed to get file contents")
	}
	if len(files) < 1 {
		return nil, errors.New("no files found")
	}
	if len(files) != files[0].SetCount {
		return nil, errors.New("incomplete file set")
	}

	sort.Sort(fileModelstruct(files))

	contents := make([][]byte, len(files))
	for i, file := range files {
		contents[i] = file.Contents
	}

	return contents, nil
}
