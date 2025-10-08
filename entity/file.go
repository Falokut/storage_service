package entity

import "io"

const (
	FilePrettyNameMetadataField      = "PrettyName"
	FileMinioMetadataPrettyNameField = "X-Amz-Meta-PrettyName"
)

type Metadata struct {
	Filename    string
	PrettyName  string
	Category    string
	ContentType string
	Size        int64
}

type UploadFileRequest struct {
	Filename      string
	PrettyName    string
	Category      string
	Pending       bool
	ContentReader io.Reader
}

type FileToDelete struct {
	Filename string
	Category string
}
