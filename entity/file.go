package entity

type File struct {
	Filename    string
	Category    string
	ContentType string
	Content     []byte
	Size        int64
}
