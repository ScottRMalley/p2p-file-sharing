package protocol

import (
	"github.com/scottrmalley/p2p-file-challenge/model"
	"sort"
	"sync"
)

type pendingFiles []model.File

func (p pendingFiles) Len() int {
	return len(p)
}
func (p pendingFiles) Less(i, j int) bool {
	return p[i].Metadata.FileNumber < p[j].Metadata.FileNumber
}

func (p pendingFiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type pendingFileCache struct {
	// map of set id to list of pending files
	pendingFiles map[string][]model.File
	mu           sync.Mutex
}

func (c *pendingFileCache) AddFile(file model.File) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pendingFiles[file.Metadata.SetId] = append(
		c.pendingFiles[file.Metadata.SetId], file,
	)
}

func (c *pendingFileCache) IsComplete(setId string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// every file in the set has access to the full set count, so we can
	// get away with just checking the first file
	return len(c.pendingFiles[setId]) == c.pendingFiles[setId][0].Metadata.SetCount
}

func (c *pendingFileCache) Remove(setId string) []model.File {
	c.mu.Lock()
	defer c.mu.Unlock()

	files := c.pendingFiles[setId]
	delete(c.pendingFiles, setId)

	sort.Sort(pendingFiles(files))
	return files
}
