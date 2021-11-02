package dto

import "io"

type FileItemFilterDto struct {
	Level       int
	Md5Seed     string
	FileName    string
	FilePointer *io.ReadCloser
	// FilePointer interface{}
}
