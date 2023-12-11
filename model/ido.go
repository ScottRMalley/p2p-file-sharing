package model

type FileMetadata struct {
	SetId      string `json:"set_id"`
	SetCount   int    `json:"set_count"`
	FileNumber int    `json:"file_number"`
}

type File struct {
	Metadata FileMetadata `json:"metadata"`
	Contents []byte       `json:"contents"`
}
