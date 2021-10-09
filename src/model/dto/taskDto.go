package dto

import (
	"net/http"
	"sync"
)

type TaskDto struct {
	WorkType    string
	Refers      []string
	Cookies     []*http.Cookie
	DownloadUrl string
	Name        string
	FileName    string
	FilePaths   []string
	Lan         string
	Subtitles   string
	Wg          *sync.WaitGroup
	StoreFunc   func([]*SubtitlesIndexDto)
}
