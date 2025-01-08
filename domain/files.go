package domain

type UploadFileRequest struct {
	Category string `validate:"required"`
	Filename string
	Content  []byte `validate:"required"`
}

type UploadFileResponse struct {
	Filename string
}

type FileRequest struct {
	Filename string `validate:"required"`
	Category string `validate:"required"`
}

type FileExistResponse struct {
	FileExist bool
}
