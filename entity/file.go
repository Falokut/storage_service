package entity

type File struct {
	Metadata Metadata
	Content  []byte
}

type Metadata struct {
	Filename    string
	Category    string
	ContentType string
	Size        int64
}
