package dto

import (
	"net/http"
	"strings"
	"sync"
	"unicode"
)

type TaskDto struct {
	SearchKeyword string //搜索的关键词
	WorkType      string
	Refers        []string
	Cookies       []*http.Cookie
	Wg            *sync.WaitGroup

	Name          string //页面显示的字幕名称
	Lan           string //字幕采用的语言
	SubtitlesType string //字幕文件格式 如SSA

	DownloadUrl         string              //下载链接
	DownloadUrlFileName string              //下载链接带的文件名称（比如1.rar）
	SubtitlesFiles      []*SubtitlesFileDto //解压后存储的文件列表
	StoreFunc           func(taskDto *TaskDto)
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
