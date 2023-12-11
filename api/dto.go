package api

type PostFilesRequest struct {
	Files []string `json:"files" validate:"required,max=1000"`
}

type PostFilesResponse struct {
	Success bool   `json:"success"`
	SetId   string `json:"setId"`
	Root    string `json:"hash"`
}

type PostFileRequest struct {
	Content  string `json:"content" validate:"required"`
	SetCount int    `json:"setCount" validate:"required"`

	SetId string `path:"setId"`
	Index int    `path:"index"`
}

type PostFileResponse struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash"`
}

type GetFileRequest struct {
	SetId string `path:"setId" validate:"required"`
	Index int    `path:"index" validate:"required"`
}

type GetFileResponse struct {
	File  string        `json:"file"`
	Proof ProofResponse `json:"proof"`
}

type ProofResponse struct {
	Proof []string `json:"proof"`
	Index uint64   `json:"index"`
}
