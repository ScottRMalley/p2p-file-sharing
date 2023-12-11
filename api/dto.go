package api

type FilesInput struct {
	Files []string `json:"files" validate:"required"`
}

type FilesOutput struct {
	Success bool   `json:"success"`
	SetId   string `json:"setId"`
	Root    string `json:"hash"`
}

type GetFileInput struct {
	SetId string `path:"setId" validate:"required"`
	Index int    `path:"index" validate:"required"`
}

type GetFileOutput struct {
	File  string        `json:"file"`
	Proof ProofResponse `json:"proof"`
}

type ProofResponse struct {
	Proof []string `json:"proof"`
	Index uint64   `json:"index"`
}
