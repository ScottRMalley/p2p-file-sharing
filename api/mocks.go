package api

import (
	"context"

	"github.com/scottrmalley/p2p-file-challenge/model"
)

type persistenceMock struct {
	files map[string][]model.File
}

func newPersistenceMock() *persistenceMock {
	return &persistenceMock{
		files: make(map[string][]model.File),
	}
}

func (p *persistenceMock) File(setId string, index int) (model.File, error) {
	return p.files[setId][index], nil
}

func (p *persistenceMock) Files(setId string) ([][]byte, error) {
	var out [][]byte
	for _, file := range p.files[setId] {
		out = append(out, file.Contents)
	}
	return out, nil
}

func (p *persistenceMock) SaveFile(file model.File) error {
	p.files[file.Metadata.SetId] = append(p.files[file.Metadata.SetId], file)
	return nil
}

func (p *persistenceMock) Write(_ context.Context, _ model.File) error {
	return nil
}
