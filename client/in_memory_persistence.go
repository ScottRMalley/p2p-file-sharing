package client

import "github.com/pkg/errors"

var ErrNotFound = errors.New("not found")

type InMemoryPersistence struct {
	fileSets map[string][]byte
	setSizes map[string]int
}

func NewInMemoryPersistence() *InMemoryPersistence {
	return &InMemoryPersistence{
		fileSets: make(map[string][]byte),
		setSizes: make(map[string]int),
	}
}

func (p *InMemoryPersistence) SetFileSet(setId string, root []byte, count int) error {
	p.fileSets[setId] = root
	p.setSizes[setId] = count
	return nil
}

func (p *InMemoryPersistence) FileSet(setId string) ([]byte, int, error) {
	root, ok := p.fileSets[setId]
	if !ok {
		return nil, 0, ErrNotFound
	}
	count, ok := p.setSizes[setId]
	if !ok {
		return nil, 0, ErrNotFound
	}
	return root, count, nil
}

func (p *InMemoryPersistence) Sets() ([]string, error) {
	var out []string
	for k := range p.fileSets {
		out = append(out, k)
	}
	return out, nil
}
