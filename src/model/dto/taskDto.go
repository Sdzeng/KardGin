package dto

import (
	"net/http"
	"strings"
	"sync"
	"unicode"
)

type TaskDto struct {
	WorkType      string
	Refers        []string
	Cookies       []*http.Cookie
	DownloadUrl   string
	Name          string
	FileName      string
	FilePathDtos  []*FilePathDto
	Lan           string
	Subtitles     string //字幕文件格式 如SSA
	SearchKeyword string
	Wg            *sync.WaitGroup
	StoreFunc     func(taskDto *TaskDto)
}

func (dto *TaskDto) ContainsKeyword(work string) bool {
	if len(dto.SearchKeyword) <= 0 {
		return false
	}

	kws := strings.FieldsFunc(dto.SearchKeyword, unicode.IsSpace) //strings.Split(dto.SearchKeyword, " ")
	for _, kw := range kws {
		if !strings.Contains(work, kw) {
			return false
		}
	}
	return true
}
