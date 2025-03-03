package entity

import "io"

type Metadata struct {
	Filename    string
	Category    string
	ContentType string
	Size        int64
}

type UploadFileRequest struct {
	Filename      string
	Category      string
	ContentReader io.Reader
}
