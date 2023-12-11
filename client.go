package p2p_file_challenge

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/scottrmalley/p2p-file-challenge/api"
	"github.com/scottrmalley/p2p-file-challenge/proof"
)

type ClientPersistence interface {
	SetFileSet(setId string, root []byte, count int) error
	FileSet(setId string) ([]byte, int, error)
	Sets() ([]string, error)
}

type Client struct {
	persistence ClientPersistence
	apiClient   *api.Client
}

func NewClient(persistence ClientPersistence, apiClient *api.Client) *Client {
	return &Client{
		persistence: persistence,
		apiClient:   apiClient,
	}
}

func (c *Client) PostFiles(files [][]byte) error {
	out, err := c.apiClient.PostFiles(
		&api.FilesInput{
			Files: func() []string {
				out := make([]string, len(files))
				for i, file := range files {
					out[i] = hexutil.Encode(file)
				}
				return out
			}(),
		},
	)
	if err != nil {
		return err
	}

	root, err := proof.Root(files)
	if err != nil {
		return err
	}
	if hexutil.Encode(root) != out.Root {
		return errors.Errorf("root mismatch: expected %s, server returned %s", root, out.Root)
	}
	if err := c.persistence.SetFileSet(out.SetId, root, len(files)); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetFile(setId string, index int) ([]byte, error) {
	root, count, err := c.persistence.FileSet(setId)
	if err != nil {
		return nil, err
	}
	if count < index {
		return nil, errors.Errorf("index %d out of range for file set %s", index, setId)
	}
	out, err := c.apiClient.GetFile(
		setId,
		index,
	)
	if err != nil {
		return nil, err
	}
	hashes, position, err := decodeProofResponse(out.Proof)
	if err != nil {
		return nil, err
	}
	file, err := hexutil.Decode(out.File)
	if err != nil {
		return nil, err
	}
	if success, err := proof.Verify(file, hashes, position, root); err != nil {
		return nil, errors.Wrap(err, "failed to verify proof")
	} else if !success {
		return nil, errors.New("proof verification failed")
	}
	return file, nil
}

func (c *Client) Sets() ([]string, error) {
	return c.persistence.Sets()
}

func (c *Client) SetSize(setId string) (int, error) {
	_, count, err := c.persistence.FileSet(setId)
	return count, err
}

func decodeProofResponse(in api.ProofResponse) (hashes [][]byte, index uint64, err error) {
	hashes = make([][]byte, len(in.Proof))
	for i, hash := range in.Proof {
		hashes[i], err = hexutil.Decode(hash)
		if err != nil {
			return nil, 0, err
		}
	}
	return hashes, in.Index, nil
}
